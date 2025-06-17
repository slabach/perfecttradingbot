package types

import "time"

type TickData struct {
	Symbol    string
	Open      float64
	Close     float64
	High      float64
	Low       float64
	Volume    float64
	AvgVolume float64
	ATR       float64
	Timestamp int64
	Last      float64
}

type Order struct {
	ID          int
	Symbol      string
	Side        int // "buy" = 0 or "sell" = 1
	Type        int // "market" = 2, "limit" = 1, "stopLimit" = 4, etc.
	Price       float64
	StopPrice   float64
	Qty         int
	TIF         string // Time in force, e.g., "GTC", "IOC"
	TargetPrice float64
	CustomTag   string
}

type Bar struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int
}

type Position struct {
	Symbol   string  // Contract ID, e.g. "CON.F.US.NQ.Z24"
	Quantity int     // Positive = long, negative = short
	Price    float64 // Optional: entry price
}

type Fill struct {
	OrderID int     // The ID of the order that was filled
	Price   float64 // The price at which the fill occurred
	Size    int     // Number of contracts filled
	// Optional:
	// Timestamp int64  // When the fill happened (from market data)
	// Side      string // "BUY" or "SELL", if available from fill event
}
