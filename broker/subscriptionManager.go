package broker

//
//import (
//	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
//	"log"
//	"perfectTradingBot/types"
//	"sync"
//	"time"
//)
//
//package subscription
//
//import (
//"log"
//"sync"
//"time"
//
//"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
//)
//
//func NewSubscriptionManager(tradeStream stream.StocksClient, max int) *types.SubscriptionManager {
//	return &types.SubscriptionManager{
//		tradeStream:        tradeStream,
//		activeTradeSubs:    make(map[string]bool),
//		lastSeen:           make(map[string]time.Time),
//		maxTradeSubs:       max,
//		tradeSubCandidates: make(chan string, 100),
//	}
//}
//
//func (s *types.SubscriptionManager) Start() {
//	go func() {
//		for sym := range s.tradeSubCandidates {
//			s.mu.Lock()
//
//			// mark latest activity
//			s.lastSeen[sym] = time.Now()
//
//			if !s.activeTradeSubs[sym] {
//				if len(s.activeTradeSubs) >= s.maxTradeSubs {
//					s.evictOldestSub()
//				}
//				err := s.tradeStream.SubscribeToTrades(sym)
//				if err == nil {
//					s.activeTradeSubs[sym] = true
//					log.Printf("Subscribed to trades: %s", sym)
//				} else {
//					log.Printf("Failed to subscribe to trades for %s: %v", sym, err)
//				}
//			}
//
//			s.mu.Unlock()
//		}
//	}()
//
//	// periodically remove stale subs
//	go func() {
//		ticker := time.NewTicker(1 * time.Minute)
//		for range ticker.C {
//			s.cleanupInactiveSubs(10 * time.Minute)
//		}
//	}()
//}
//
//func (s *types.SubscriptionManager) evictOldestSub() {
//	for sym := range s.activeTradeSubs {
//		_ = s.tradeStream.UnsubscribeFromTrades(sym)
//		delete(s.activeTradeSubs, sym)
//		delete(s.lastSeen, sym)
//		log.Printf("Unsubscribed from trades: %s", sym)
//		break
//	}
//}
//
//func (s *types.SubscriptionManager) cleanupInactiveSubs(maxAge time.Duration) {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	now := time.Now()
//	for sym := range s.activeTradeSubs {
//		if last, ok := s.lastSeen[sym]; ok && now.Sub(last) > maxAge {
//			_ = s.tradeStream.UnsubscribeFromTrades(sym)
//			delete(s.activeTradeSubs, sym)
//			delete(s.lastSeen, sym)
//			log.Printf("Auto-unsubscribed from stale symbol: %s", sym)
//		}
//	}
//}
//
//func (s *types.SubscriptionManager) RequestTradeSub(sym string) {
//	s.mu.Lock()
//	s.lastSeen[sym] = time.Now()
//	s.mu.Unlock()
//
//	select {
//	case s.tradeSubCandidates <- sym:
//	default:
//		log.Printf("Trade subscription channel full, skipping %s", sym)
//	}
//}
