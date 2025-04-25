package broker

import (
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
	"log"
	"math"
	"perfectTradingBot/global"
	"perfectTradingBot/types"
	"perfectTradingBot/utils"
	"time"
)

func HandleTrade(trade stream.Trade) {
	global.StateLock.Lock()
	defer global.StateLock.Unlock()
	sym := trade.Symbol
	price := trade.Price
	now := trade.Timestamp.In(time.FixedZone("EST", -5*60*60))

	// === EXIT LOGIC with real-time trades ===
	if pos, ok := global.Positions[sym]; ok && pos.Shares > 0 {
		peak := math.Max(pos.PeakPrice, price)
		if price > pos.PeakPrice {
			pos.PeakPrice = price
			global.Positions[sym] = pos
		}
		drawdown := (peak - price) / peak
		profit := (price - pos.CostBasis) / pos.CostBasis
		heldFor := time.Since(pos.EntryTime)

		if drawdown > 0.04 || (profit >= 0.12 && price < pos.PeakPrice*0.9975) || profit <= -0.03 || heldFor > 15*time.Minute {
			ExecuteSell(sym, price)
			return
		}
	}

	// === ENTRY LOGIC ===
	if !global.ReclaimConfirmed[sym] {
		return
	}

	if t, traded := global.LastTradeTime[sym]; traded && t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return
	}

	prices := map[string]float64{sym: price}
	global.MacdStates = utils.RunMACDPool(prices, global.MacdStates, &global.StateLock)

	macd := global.MacdStates[sym]
	if macd == nil {
		macd = &types.MACDState{}
		global.MacdStates[sym] = macd
	}
	prevMACD := macd.LastMACD
	//updateMACD(macd, price)

	hist := macd.LastMACD - macd.LastSignal
	slope := macd.LastMACD - prevMACD

	if math.Abs(hist) >= global.MinHist && math.Abs(slope) >= global.MinSlope {
		if global.RequireTrend {
			if !utils.PriceAboveTrend(sym, price, global.LongEMAPeriod) {
				return
			}
		}
		if global.UseRSIFilter {
			if !utils.RsiWithinRange(sym, global.RsiMin, global.RsiMax) {
				return
			}
		}

		if global.DayTradeCount >= 3 {
			sharesToBuy := int(1000 / price)
			global.Positions[sym] = types.Position{Shares: sharesToBuy, CostBasis: price, EntryTime: now, PeakPrice: price}
			log.Printf("PAPER BUY (TRADE) %s: %d shares @ %.2f", sym, sharesToBuy, price)
			return
		}

		log.Printf("REAL-TIME BUY TRIGGERED: %s @ %.2f", sym, price)
		ExecuteBuy(sym, price)
		global.LastTradeTime[sym] = now
		global.AnticipatoryFlags[sym] = false
		global.ReclaimConfirmed[sym] = false
	}
}
