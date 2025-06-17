package strategy

import (
	"math"
	"perfectTradingBot/types"
	"time"
)

type BreakoutStrategy struct {
	lastTradeTime time.Time
}

func NewBreakoutStrategy() *BreakoutStrategy {
	return &BreakoutStrategy{}
}

func (s *BreakoutStrategy) ShouldTrade(tick types.TickData) bool {
	if tick.Volume > tick.AvgVolume*1.5 && math.Abs(tick.Close-tick.Open) > tick.ATR {
		if time.Since(s.lastTradeTime).Minutes() > 5 {
			return true
		}
	}
	return false
}

func (s *BreakoutStrategy) GenerateOrder(tick types.TickData) types.Order {
	s.lastTradeTime = time.Now()

	entry := tick.Close
	stop := entry - (5 * tick.ATR)    // dynamic stop-loss
	target := entry + (10 * tick.ATR) // dynamic take-profit

	return types.Order{
		Side:        0, // BUY
		Qty:         5,
		StopPrice:   stop,
		TargetPrice: target,
		Type:        2, // MARKET
		CustomTag:   "Order-Placed",
	}

}

func (s *BreakoutStrategy) Name() string {
	return "Volume+ATR Breakout"
}
