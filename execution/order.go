package execution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"perfectTradingBot/marketdata"
	"perfectTradingBot/types"
	"strconv"
	_ "strings"
	_ "time"
)

type orderPayload struct {
	AccountID     int64    `json:"accountId"`
	Symbol        *string  `json:"symbol"`
	Size          *int     `json:"size"`
	OrderType     *string  `json:"orderType"`
	Price         *float64 `json:"price,omitempty"`
	ContractId    *string  `json:"contractId"`
	Side          *int     `json:"side"`
	Type          *int     `json:"type"`
	LimitPrice    *float64 `json:"limitPrice,omitempty"`
	StopPrice     *float64 `json:"stopPrice,omitempty"`
	TrailPrice    *float64 `json:"trailPrice,omitempty"`
	CustomTag     *string  `json:"customTag,omitempty"`
	LinkedOrderID *int     `json:"linkedOrderId,omitempty"`
	OrderID       *int     `json:"orderId,omitempty"`
}

type placeOrderResponse struct {
	OrderID      int    `json:"orderId"`
	Success      bool   `json:"success"`
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

func SubmitOrder(order types.Order) (*int, error) {
	accountIDStr := os.Getenv("PROJECTX_ACCOUNT_ID")
	token := os.Getenv("PROJECTX_SESSION_TOKEN")
	if accountIDStr == "" || token == "" {
		return nil, fmt.Errorf("missing account ID or auth token")
	}

	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID: %v", err)
	}

	contractId := os.Getenv("PROJECTX_CON_ID")
	payload := orderPayload{
		AccountID:     accountID,
		ContractId:    &contractId,
		Type:          &order.Type,
		Side:          &order.Side,
		Size:          &order.Qty,
		CustomTag:     &order.CustomTag,
		LinkedOrderID: order.LinkedOrderID,
	}
	if order.Type != 2 {
		payload.LimitPrice = order.TargetPrice
		payload.StopPrice = order.StopPrice
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "https://api.topstepx.com/api/Order/place", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("order rejected: %s", resp.Status)
	}
	respBody, err := io.ReadAll(resp.Body)
	log.Printf("Order response: %s", string(respBody))
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	if len(respBody) == 0 {
		return nil, fmt.Errorf("empty response from server (EOF)")
	}

	var parsed placeOrderResponse
	err = json.Unmarshal(respBody, &parsed)
	if err != nil {
		return nil, err
	}

	if !parsed.Success {
		return nil, fmt.Errorf("order rejected: code=%d message=%s", parsed.ErrorCode, parsed.ErrorMessage)
	}

	return &parsed.OrderID, nil
}

func CancelOrder(orderID int) {
	url := fmt.Sprintf("https://api.topstepx.com/api/Order/cancel")
	accountIDStr := os.Getenv("PROJECTX_ACCOUNT_ID")
	if accountIDStr == "" {
		log.Println("missing account ID")
		return
	}

	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		return
	}

	payload := orderPayload{
		AccountID: accountID,
		OrderID:   &orderID,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("PROJECTX_SESSION_TOKEN"))
	req.Header.Set("Accept", "text/plain")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("cancel error: %v", err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("cancel failed for order %d", orderID)
	}
}

func WaitForFill(order types.Order) *types.Order {
	defer marketdata.RemoveOrderListener(order.ID)
	// Use SignalR real-time updates instead of polling
	filled := marketdata.BlockUntilOrderFilled(order)
	if filled {
		return &order
	}
	return nil
}
