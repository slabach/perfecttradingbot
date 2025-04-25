package types

type MACDJob struct {
	Symbol string
	Price  float64
}

type MACDResult struct {
	Symbol string
	State  *MACDState
}
