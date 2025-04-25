package types

import "time"

type MACDState struct {
	FastEMA     float64
	SlowEMA     float64
	SignalEMA   float64
	LastMACD    float64
	LastSignal  float64
	Initialized bool
	LastCrossUp bool
}

type AggregatedBar struct {
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
	Start  time.Time
}

type VWAPState struct {
	TotalPV  float64
	TotalVol float64
}
