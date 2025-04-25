package utils

import "perfectTradingBot/global"

func RsiWithinRange(sym string, min, max float64) bool {
	windowSize := 14
	priceHistory := global.RsiHistory[sym]
	if len(priceHistory) < windowSize+1 {
		return true // not enough data to evaluate RSI
	}

	gains := 0.0
	losses := 0.0
	for i := 1; i <= windowSize; i++ {
		diff := priceHistory[len(priceHistory)-i] - priceHistory[len(priceHistory)-i-1]
		if diff > 0 {
			gains += diff
		} else {
			losses -= diff
		}
	}

	if gains+losses == 0 {
		return true // no movement
	}

	avgGain := gains / float64(windowSize)
	avgLoss := losses / float64(windowSize)

	if avgLoss == 0 {
		return true // prevent division by zero
	}

	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1 + rs))

	return rsi >= min && rsi <= max
}

func PriceAboveTrend(sym string, price float64, period int) bool {
	history := global.RsiHistory[sym]
	if len(history) < period {
		return true
	}

	ema := history[len(history)-period]
	alpha := 2.0 / float64(period+1)
	for i := len(history) - period + 1; i < len(history); i++ {
		ema = alpha*history[i] + (1-alpha)*ema
	}
	return price > ema
}
