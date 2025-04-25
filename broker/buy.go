package broker

import (
	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/shopspring/decimal"
	"log"
	"math"
	"perfectTradingBot/global"
	"perfectTradingBot/types"
	"perfectTradingBot/utils"
	"time"
)

func ExecuteBuy(sym string, price float64) {
	if global.Positions[sym].Shares > 0 {
		log.Printf("already maintain position in: %s", sym)
		return
	}
	sharesToBuy := utils.CalculateOrderSize(price)
	if sharesToBuy < 1 {
		log.Printf("not enough buying power to purchase %s @ %.2f\n", sym, price)
		return
	}

	//qty := decimal.NewFromInt(int64(sharesToBuy))
	//_, err := alpacaClient.PlaceOrder(alpaca.PlaceOrderRequest{
	//	Symbol:      sym,
	//	Qty:         &qty,
	//	Side:        alpaca.Buy,
	//	Type:        alpaca.Market,
	//	TimeInForce: alpaca.GTC,
	//})
	qty := decimal.NewFromInt(int64(sharesToBuy))
	limitPrice := decimal.NewFromFloat(math.Round(price*1.001*100) / 100)
	_, err := global.AlpacaClient.PlaceOrder(alpaca.PlaceOrderRequest{
		Symbol:      sym,
		Qty:         &qty,
		LimitPrice:  &limitPrice,
		Type:        alpaca.Limit,
		TimeInForce: alpaca.GTC,
		Side:        alpaca.Buy,
	})
	if err != nil {
		log.Printf("live order error: %v\n", err)
		return
	}

	global.Positions[sym] = types.Position{Shares: sharesToBuy, CostBasis: price, EntryTime: time.Now(), PeakPrice: price}
	log.Printf("BUY %s: %d shares @ %.2f | Cash: %.2f\n", sym, sharesToBuy, price, price*float64(sharesToBuy))
	global.CashBalance = price * float64(sharesToBuy)
}
