package risk

import (
	"log"
	//"perfectTradingBot/execution"
	"github.com/perfecttradingbot/types"
	"sync"
	"time"
)

type Manager struct {
	dailyTrades   int
	dailyLoss     float64
	maxTrades     int
	maxDailyLoss  float64
	tradeCooldown time.Duration
	lastTradeTime time.Time
	mu            sync.Mutex
	positions     map[string]types.Position
	orders        map[int]types.Order
	fills         []types.Fill
}

func NewManager() *Manager {
	return &Manager{
		maxTrades:     5,
		maxDailyLoss:  300.0, // Example max daily loss in USD
		tradeCooldown: 1 * time.Minute,
	}
}

func (m *Manager) AllowsTrade() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.dailyTrades >= m.maxTrades {
		log.Println("Trade blocked: daily trade limit reached")
		return false
	}
	if m.dailyLoss <= -m.maxDailyLoss {
		log.Println("Trade blocked: max daily loss reached")
		return false
	}
	if time.Since(m.lastTradeTime) < m.tradeCooldown {
		log.Println("Trade blocked: cooldown active")
		return false
	}
	if IsInRestrictedWindow() {
		log.Println("Trading blocked during restricted window (3:10 PM to 5:00 PM CT)")
		return false
	}
	for _, pos := range m.positions {
		if pos.Quantity != 0 {
			return false
		}
	}
	return true
}

func (m *Manager) RegisterTrade(order types.Order) {
	m.dailyTrades++
	m.lastTradeTime = time.Now()
	// In a real scenario, we'd fetch execution price + unrealized PnL tracking
	log.Printf("Registered trade: %+v\n", order)
}

func (m *Manager) UpdateOrder(data map[string]interface{}) {
	log.Printf("[Risk] Order update received: %+v\n", data)
	// Handle order status updates (submitted, filled, canceled)
	m.mu.Lock()
	defer m.mu.Unlock()
	id := int(data["orderId"].(float64))
	m.orders[id] = types.Order{ID: id}
}

func (m *Manager) UpdatePosition(data map[string]interface{}) {
	log.Printf("[Risk] Position update received: %+v\n", data)
	// Update position tracking (size, entry price, PnL, etc.)
	m.mu.Lock()
	defer m.mu.Unlock()
	symbol := data["contractId"].(string)
	qty := int(data["position"].(float64))
	m.positions[symbol] = types.Position{Symbol: symbol, Quantity: qty}
}

func (m *Manager) UpdateFill(data map[string]interface{}) {
	log.Printf("[Risk] Fill update received: %+v\n", data)
	// Update realized PnL, match fill to order, etc.
	m.mu.Lock()
	defer m.mu.Unlock()
	fill := types.Fill{
		OrderID: int(data["orderId"].(float64)),
		Price:   data["price"].(float64),
		Size:    int(data["size"].(float64)),
	}
	m.fills = append(m.fills, fill)
}

func IsInRestrictedWindow() bool {
	loc, _ := time.LoadLocation("America/Chicago")
	now := time.Now().In(loc)

	start := time.Date(now.Year(), now.Month(), now.Day(), 15, 10, 0, 0, loc) // 3:10 PM CT
	end := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, loc)    // 5:00 PM CT

	return now.After(start) && now.Before(end)
}

//func (m *Manager) FlattenAllPositions() {
//	for _, pos := range m.positions {
//		if pos.Quantity != 0 {
//			log.Printf("[Risk] Flattening position: %+v", pos)
//
//			side := 1 // SELL
//			if pos.Quantity < 0 {
//				side = 0 // BUY
//			}
//
//			order := types.Order{
//				Symbol: pos.Symbol,
//				Qty:    intAbs(pos.Quantity),
//				Side:   side,
//				Type:   2, // MARKET
//			}
//
//			err := execution.SubmitOrder(order)
//			if err != nil {
//				log.Printf("[Risk] Failed to flatten position %s: %v", pos.Symbol, err)
//				continue
//			}
//		}
//	}
//}

func intAbs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
