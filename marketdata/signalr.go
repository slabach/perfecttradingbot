package marketdata

import (
	"context"
	"fmt"
	"log"
	"os"
	"perfectTradingBot/auth"
	"perfectTradingBot/risk"
	"perfectTradingBot/types"
	"strconv"
	"sync"
	"time"

	"github.com/philippseith/signalr"
)

type signalrNoOpLogger struct{}

func (l signalrNoOpLogger) Log(keyVals ...interface{}) error {
	// No-op: discard logs
	return nil
}

var (
	orderFillMap   = make(map[int]chan bool)
	orderFillMutex sync.Mutex
)

type marketReceiver struct {
	TickChan chan types.TickData
}

type userReceiver struct {
	risk *risk.Manager
}

func (r *userReceiver) GatewayUserAccount(data map[string]interface{}) {
	log.Printf("[User Account Update] %+v", data)
}

func (r *userReceiver) GatewayUserOrder(data map[string]interface{}) {
	r.risk.UpdateOrder(data)

	orderID, ok := data["orderId"].(float64)
	status, statusOk := data["status"].(string)
	if ok && statusOk {
		if status == "Filled" || status == "PartiallyFilled" {
			orderFillMutex.Lock()
			if ch, exists := orderFillMap[int(orderID)]; exists {
				select {
				case ch <- true:
				default:
				}
			}
			orderFillMutex.Unlock()
		}
	}
}

func (r *userReceiver) GatewayUserPosition(data map[string]interface{}) {
	r.risk.UpdatePosition(data)
}

func (r *userReceiver) GatewayUserTrade(data map[string]interface{}) {
	r.risk.UpdateFill(data)
}

func (r *marketReceiver) GatewayQuote(contractID string, data map[string]interface{}) {
	//log.Printf("Quote [%s]: %+v\n", contractID, data)
}

func (r *marketReceiver) GatewayTrade(contractID string, data []map[string]interface{}) {
	for _, trade := range data {
		price, _ := trade["price"].(float64)
		volume, _ := trade["volume"].(float64)
		timestamp := time.Now().UnixMilli()

		tick := types.TickData{
			Symbol:    contractID,
			Close:     price,
			Volume:    volume,
			Timestamp: timestamp,
		}
		r.TickChan <- tick
	}
}

func (r *marketReceiver) GatewayDepth(contractID string, data []map[string]interface{}) {
	//for _, entry := range data {
	//	//log.Printf("Depth [%s]: %+v\n", contractID, entry)
	//}
}

func ConnectToMarketHub(tickChan chan types.TickData) {
	token := auth.GetSessionToken()
	url := fmt.Sprintf("https://rtc.topstepx.com/hubs/market?access_token=%s", token)

	receiver := &marketReceiver{TickChan: tickChan}

	ctx := context.Background()
	conn, err := signalr.NewHTTPConnection(ctx, url)
	if err != nil {
		log.Fatalf("failed to create HTTP connection: %v", err)
	}

	client, err := signalr.NewClient(ctx,
		signalr.WithConnection(conn),
		signalr.TransferFormat(signalr.TransferFormatText),
		signalr.WithReceiver(receiver),
		signalr.Logger(signalrNoOpLogger{}, false),
	)
	if err != nil {
		log.Fatalf("failed to create SignalR client: %v", err)
	}

	client.Start()

	contractID := os.Getenv("PROJECTX_CON_ID")
	go func() {
		<-time.After(1 * time.Second)
		client.Send("SubscribeContractQuotes", contractID)
		client.Send("SubscribeContractTrades", contractID)
		// client.Send("SubscribeContractMarketDepth", contractID)
	}()

	go func() {
		for {
			now := time.Now().In(time.FixedZone("CT", -6*3600))
			flattenTime := time.Date(now.Year(), now.Month(), now.Day(), 15, 9, 0, 0, now.Location())
			delta := time.Until(flattenTime)
			if delta < 0 {
				flattenTime = flattenTime.Add(24 * time.Hour)
				delta = time.Until(flattenTime)
			}
			time.Sleep(delta)
			log.Println("Auto-flattening triggered at 3:09 PM CT")
			//risk.FlattenAllPositions()
			time.Sleep(1 * time.Minute) // avoid duplicate trigger
		}
	}()
}

func ConnectToUserHub() {
	token := auth.GetSessionToken()
	url := fmt.Sprintf("https://rtc.topstepx.com/hubs/user?access_token=%s", token)

	accountIDStr := os.Getenv("PROJECTX_ACCOUNT_ID")
	if accountIDStr == "" {
		log.Println("missing account ID")
		return
	}

	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		return
	}

	receiver := &userReceiver{}

	ctx := context.Background()
	conn, err := signalr.NewHTTPConnection(ctx, url)
	if err != nil {
		log.Fatalf("failed to create user hub connection: %v", err)
	}

	client, err := signalr.NewClient(ctx,
		signalr.WithConnection(conn),
		signalr.TransferFormat(signalr.TransferFormatText),
		signalr.WithReceiver(receiver),
		signalr.Logger(signalrNoOpLogger{}, false),
	)
	if err != nil {
		log.Fatalf("failed to create user hub client: %v", err)
	}

	client.Start()

	go func() {
		<-time.After(1 * time.Second)
		client.Send("SubscribeAccounts")
		client.Send("SubscribeOrders", accountID)
		client.Send("SubscribePositions", accountID)
		client.Send("SubscribeTrades", accountID)
	}()
}

func BlockUntilOrderFilled(order types.Order) bool {
	ch := make(chan bool, 1)
	orderFillMutex.Lock()
	orderFillMap[order.ID] = ch
	orderFillMutex.Unlock()
	timer := time.NewTimer(10 * time.Second)
	select {
	case <-ch:
		log.Printf("Order %d filled.", order.ID)
		return true
	case <-timer.C:
		log.Printf("Timeout waiting for fill on order %d", order.ID)
		return false
	}
}

func RemoveOrderListener(orderID int) {
	orderFillMutex.Lock()
	delete(orderFillMap, orderID)
	orderFillMutex.Unlock()
}
