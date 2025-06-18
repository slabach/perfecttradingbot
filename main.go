// File: main.go
package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"perfectTradingBot/auth"
	"perfectTradingBot/execution"
	"perfectTradingBot/marketdata"
	"perfectTradingBot/risk"
	"perfectTradingBot/strategy"
	"perfectTradingBot/types"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	log.Println("Starting ProjectX NQ Bot...")

	// Authenticate and get session token
	if err := auth.LoginWithKey(); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	// Search for active accounts
	accountID, err := auth.SearchActiveAccount()
	if err != nil {
		log.Fatalf("Account search failed: %v", err)
	}
	err = os.Setenv("PROJECTX_ACCOUNT_ID", accountID)
	if err != nil {
		return
	}
	go marketdata.ConnectToUserHub()

	// Example: fetch historical bars
	//bars, err := marketdata.FetchHistoricalBars(os.Getenv("PROJECTX_CON_ID"))
	//if err != nil {
	//	log.Fatalf("Bar fetch failed: %v", err)
	//}
	//log.Printf("Retrieved %d bars\n", len(bars))

	riskManager := risk.NewManager()

	var ticks <-chan types.TickData

	tickChan := make(chan types.TickData, 100)
	go marketdata.ConnectToMarketHub(tickChan)
	ticks = tickChan

	s := strategy.NewStrategyManager()

	for tick := range ticks {
		if s.ShouldTrade(tick) && riskManager.AllowsTrade() {
			order := s.GenerateOrder(tick)
			mainOrderId, err := execution.SubmitOrder(order)
			if err != nil {
				log.Printf("Failed to submit order: %v", err)
				continue
			}

			var stopOrderId, targetOrderId *int
			// submit attached stop-loss if defined
			if order.StopPrice != nil && mainOrderId != nil {
				if *order.StopPrice > 0 {
					stopOrder := types.Order{
						Side:          1, // SELL
						Type:          4, // STOP
						Qty:           order.Qty,
						StopPrice:     order.StopPrice,
						CustomTag:     "stop-loss",
						LinkedOrderID: mainOrderId,
					}
					stopOrderId, err = execution.SubmitOrder(stopOrder)
					if err != nil {
						fmt.Printf("\nFailed to submit target order: %v\n", err)
					} else if stopOrderId != nil {
						stopOrder.ID = *stopOrderId
					}
				}
			}

			// submit attached take-profit if defined
			if order.TargetPrice != nil && mainOrderId != nil {
				if *order.TargetPrice > 0 {
					targetOrder := types.Order{
						Side:          1, // SELL
						Type:          1, // LIMIT
						Qty:           order.Qty,
						TargetPrice:   order.TargetPrice,
						CustomTag:     "take-profit",
						LinkedOrderID: mainOrderId,
					}
					targetOrderId, err = execution.SubmitOrder(targetOrder)
					if err != nil {
						fmt.Printf("\nFailed to submit stop order: %v\n", err)
					} else if targetOrderId != nil {
						targetOrder.ID = *targetOrderId
					}
				}
			}

			// monitor for order fulfillment
			go func() {
				filled := execution.WaitForFill(order)
				if filled != nil {
					if stopOrderId != nil && *stopOrderId > 0 {
						execution.CancelOrder(*stopOrderId)
					}
					if targetOrderId != nil && *targetOrderId > 0 {
						execution.CancelOrder(*targetOrderId)
					}
				}
			}()

			riskManager.RegisterTrade(order)
		}
	}
}
