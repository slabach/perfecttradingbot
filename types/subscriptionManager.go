package types

import (
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
	"sync"
	"time"
)

type SubscriptionManager struct {
	mu                 sync.Mutex
	activeTradeSubs    map[string]bool
	lastSeen           map[string]time.Time
	maxTradeSubs       int
	tradeStream        stream.StocksClient
	tradeSubCandidates chan string
}
