package broker

import (
	"math"
	"perfectTradingBot/global"
	"perfectTradingBot/types"
	"perfectTradingBot/utils"
	"time"
)

func ProcessFiveMinBar(sym string, bar types.AggregatedBar) {
	price := bar.Close
	global.StateLock.Lock()
	defer global.StateLock.Unlock()

	if !global.AnticipatoryFlags[sym] {
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

	pos, ok := global.Positions[sym]
	if ok && pos.Shares > 0 {
		peak := math.Max(pos.PeakPrice, price)
		if price > pos.PeakPrice {
			pos.PeakPrice = price
			global.Positions[sym] = pos
		}
		drawdown := (peak - price) / peak
		profit := (price - pos.CostBasis) / pos.CostBasis
		heldFor := time.Since(pos.EntryTime)
		if drawdown > 0.04 || profit >= 0.12 || profit <= -0.03 || heldFor > 30*time.Minute {
			ExecuteSell(sym, price)
			return
		}
	}

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

	hist := macd.LastMACD - macd.LastSignal
	slope := macd.LastMACD - prevMACD
	if math.Abs(hist) < global.MinHist || math.Abs(slope) < global.MinSlope {
		return
	}

	// Entry already handled in 1-min logic
}
