package marketdata

import (
	"perfectTradingBot/types"
	_ "time"
)

type BacktestProvider struct {
	ch chan types.TickData
}

func (b *BacktestProvider) TickStream() <-chan types.TickData {
	return b.ch
}

func NewBacktestProviderFromBars(bars []types.Bar) *BacktestProvider {
	bt := &BacktestProvider{
		ch: make(chan types.TickData, len(bars)),
	}
	go bt.loadFromBars(bars)
	return bt
}

func (b *BacktestProvider) loadFromBars(bars []types.Bar) {
	var history []types.TickData
	const maxHistory = 50

	for _, bar := range bars {
		tick := types.TickData{
			Timestamp: bar.Timestamp.Unix(),
			Last:      bar.Close,
			Volume:    float64(bar.Volume),
		}

		history = append(history, tick)
		if len(history) > maxHistory {
			history = history[1:]
		}

		tick.AvgVolume = averageVolume(history)
		tick.ATR = computeATR(history)
		b.ch <- tick
	}
	close(b.ch)
}

func averageVolume(history []types.TickData) float64 {
	if len(history) == 0 {
		return 0
	}
	sum := 0.0
	for _, h := range history {
		sum += h.Volume
	}
	return sum / float64(len(history))
}

func computeATR(history []types.TickData) float64 {
	if len(history) < 2 {
		return 0
	}
	sum := 0.0
	for i := 1; i < len(history); i++ {
		rangeVal := abs(history[i].Last - history[i-1].Last)
		sum += rangeVal
	}
	return sum / float64(len(history)-1)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
