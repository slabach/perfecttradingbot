package strategy

import "github.com/perfecttradingbot/types"

type Strategy interface {
	ShouldTrade(tick types.TickData) bool
	GenerateOrder(tick types.TickData) types.Order
	Name() string
}

type StrategyManager struct {
	strategies []Strategy
}

func NewStrategyManager() *StrategyManager {
	return &StrategyManager{
		strategies: []Strategy{
			NewBreakoutStrategy(),
		},
	}
}

func (s *StrategyManager) ShouldTrade(tick types.TickData) bool {
	for _, strat := range s.strategies {
		if strat.ShouldTrade(tick) {
			return true
		}
	}
	return false
}

func (s *StrategyManager) GenerateOrder(tick types.TickData) types.Order {
	for _, strat := range s.strategies {
		if strat.ShouldTrade(tick) {
			return strat.GenerateOrder(tick)
		}
	}
	return types.Order{}
}
