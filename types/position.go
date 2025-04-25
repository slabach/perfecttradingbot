package types

import "time"

type Position struct {
	Shares    int
	CostBasis float64
	EntryTime time.Time
	PeakPrice float64
}
