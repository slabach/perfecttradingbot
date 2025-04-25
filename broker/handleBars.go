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

func HandleBar(bar stream.Bar) {
	global.StateLock.Lock()
	defer global.StateLock.Unlock()
	sym := bar.Symbol
	now := bar.Timestamp.In(time.FixedZone("EST", -5*60*60))
	roundedTime := now.Truncate(5 * time.Minute)

	buf, exists := global.BarBuffer[sym]
	if !exists || buf.Start.Before(roundedTime) {
		if exists {
			ProcessFiveMinBar(sym, *buf)
		}
		global.BarBuffer[sym] = &types.AggregatedBar{
			Open:   bar.Open,
			High:   bar.High,
			Low:    bar.Low,
			Close:  bar.Close,
			Volume: float64(bar.Volume),
			Start:  roundedTime,
		}
		return
	}

	buf.High = math.Max(buf.High, bar.High)
	buf.Low = math.Min(buf.Low, bar.Low)
	buf.Close = bar.Close
	buf.Volume += float64(bar.Volume)

	price := bar.Close
	vwap := utils.UpdateVWAP(sym, price, float64(bar.Volume))

	// Volume tracking
	global.VolHistory[sym] = append(global.VolHistory[sym], float64(bar.Volume))
	if len(global.VolHistory[sym]) > 5 {
		global.VolHistory[sym] = global.VolHistory[sym][len(global.VolHistory[sym])-5:]
	}
	var avgVol float64
	for _, v := range global.VolHistory[sym] {
		avgVol += v
	}
	avgVol /= float64(len(global.VolHistory[sym]))

	// === Breakout + Pullback + Reclaim System ===
	if price > global.BreakoutHighs[sym] && price > vwap && float64(bar.Volume) > 2*avgVol {
		global.BreakoutHighs[sym] = price
		global.PullbackTracked[sym] = false
		global.AnticipatoryFlags[sym] = false
		global.ReclaimConfirmed[sym] = false
		log.Printf("BREAKOUT FLAGGED: %s at %.2f", sym, price)
	}

	if global.BreakoutHighs[sym] > 0 && price < global.BreakoutHighs[sym] && !global.PullbackTracked[sym] && price > vwap*0.98 {
		global.PullbackTracked[sym] = true
		log.Printf("PULLBACK SPOTTED: %s to %.2f", sym, price)
	}

	if global.PullbackTracked[sym] && price >= global.BreakoutHighs[sym] && price > vwap && float64(bar.Volume) > avgVol {
		global.ReclaimConfirmed[sym] = true
		global.AnticipatoryFlags[sym] = true
		log.Printf("RECLAIM CONFIRMED: %s at %.2f", sym, price)

		// Cooldown: skip if we've already traded this symbol today
		if t, ok := global.LastTradeTime[sym]; ok {
			if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
				log.Printf("SKIPPING TRADE - already traded %s today", sym)
				return
			}
		}

		// Update MACD using goroutine pool for all symbols
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
			log.Printf("BREAKOUT ENTRY CONFIRMED (1-min): %s at %.2f", sym, price)

			// Check PDT rule: skip if we've hit 3-day trades
			if global.DayTradeCount >= 3 {
				log.Printf("SKIPPING TRADE - reached max 3 day trades today")

				// Simulate paper trade to track what would've happened
				sharesToBuy := int(1000 / price)
				pos := types.Position{Shares: sharesToBuy, CostBasis: price, EntryTime: now, PeakPrice: price}
				global.Positions[sym] = pos
				log.Printf("PAPER BUY %s: %d shares @ %.2f", sym, sharesToBuy, price)
				return
			}

			ExecuteBuy(sym, price)
			global.LastTradeTime[sym] = now
			global.AnticipatoryFlags[sym] = false
		}
	}
}
