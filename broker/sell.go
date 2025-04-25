package broker

import (
	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/shopspring/decimal"
	"log"
	"math"
	"perfectTradingBot/global"
	"perfectTradingBot/types"
	"perfectTradingBot/utils"
)

func ExecuteSell(sym string, price float64) {
	// Check live position in Alpaca before trying to sell
	_, err := global.AlpacaClient.GetPosition(sym)
	if err != nil {
		log.Printf("SKIPPING SELL: no open position found for %s (likely manually closed)", sym)
		global.Positions[sym] = types.Position{} // clear local state
		global.RefreshAccountCache()             // refresh buying power
		return
	}

	pos, ok := global.Positions[sym]
	if !ok || pos.Shares <= 0 {
		return
	}

	//qty := decimal.NewFromInt(int64(pos.Shares))
	//_, err = alpacaClient.PlaceOrder(alpaca.PlaceOrderRequest{
	//	Symbol:      sym,
	//	Qty:         &qty,
	//	Side:        alpaca.Sell,
	//	Type:        alpaca.Market,
	//	TimeInForce: alpaca.GTC,
	//})
	qty := decimal.NewFromInt(int64(pos.Shares))
	limitPrice := decimal.NewFromFloat(math.Round(price*0.999*100) / 100)
	_, err = global.AlpacaClient.PlaceOrder(alpaca.PlaceOrderRequest{
		Symbol:      sym,
		Qty:         &qty,
		LimitPrice:  &limitPrice,
		Type:        alpaca.Limit,
		TimeInForce: alpaca.GTC,
		Side:        alpaca.Sell,
	})
	if err != nil {
		log.Printf("live sell error: %v\n", err)
		return
	}

	proceeds := price * float64(pos.Shares)
	pnl := proceeds - (pos.CostBasis * float64(pos.Shares))
	global.CashBalance += proceeds
	utils.LogTradeCSV(sym, pos.CostBasis, price, pos.Shares, pnl, pos.EntryTime)
	global.Positions[sym] = types.Position{}
	global.TotalProfitLoss += pnl
	accountChange := ((global.CashBalance - global.InitialBalance) / global.InitialBalance) * 100
	log.Printf("SELL %s: %d shares @ %.2f | P/L: %.2f | Cash: %.2f | Net P&L: %.2f | %% Change: %.2f%%", sym, pos.Shares, price, pnl, global.CashBalance, global.TotalProfitLoss, accountChange)
}
