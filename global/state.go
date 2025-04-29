package global

import (
	"encoding/gob"
	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"log"
	"os"
	"perfectTradingBot/types"
	"strconv"
	"sync"
	"time"
)

var (
	Symbols       []string
	IgnoreSymbols = map[string]bool{
		"SOXL": true,
		"TSLL": true,
		"F":    true,
		"SNAP": true,
		"TSLY": true,
		"SOFI": true,
		"SOXS": true,
		"RIVN": true,
		"LCID": true,
		"MSTU": true,
		"SPXS": true,
	}
	InitialBalance    = 1500.0
	TradeFundPercent  = 0.95
	UseLive           = false
	CashBalance       float64
	TotalProfitLoss   float64
	Positions         = make(map[string]types.Position)
	MacdStates        = make(map[string]*types.MACDState)
	RsiHistory        = make(map[string][]float64)
	StateLock         sync.RWMutex
	AlpacaClient      *alpaca.Client
	MarketClient      *marketdata.Client
	BarBuffer         = map[string]*types.AggregatedBar{}
	AnticipatoryFlags = map[string]bool{}
	BreakoutHighs     = map[string]float64{}
	PullbackTracked   = map[string]bool{}
	ReclaimConfirmed  = map[string]bool{}
	LastTradeTime     = map[string]time.Time{}
	CachedBuyingPower float64
	DayTradeCount     int
	VwapData          = map[string]*types.VWAPState{}
	VolHistory        = map[string][]float64{}
	MinHist           float64
	MinSlope          float64
	LongEMAPeriod     int
	RsiMin            float64
	RsiMax            float64
	RequireTrend      bool
	UseRSIFilter      bool
)

func PersistStatePeriodically() {
	for {
		time.Sleep(1 * time.Minute)
		StateLock.Lock()
		SaveStateToDisk()
		StateLock.Unlock()
	}
}

func SaveStateToDisk() {
	f, err := os.Create("state.gob")
	if err != nil {
		log.Printf("error saving state: %v", err)
		return
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	enc.Encode(CashBalance)
	enc.Encode(TotalProfitLoss)
	enc.Encode(Positions)
	enc.Encode(MacdStates)
	enc.Encode(RsiHistory)
}

func LoadStateFromDisk() {
	f, err := os.Open("state.gob")
	if err != nil {
		log.Println("No previous state found. Starting fresh.")
		return
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	dec.Decode(CashBalance)
	dec.Decode(TotalProfitLoss)
	dec.Decode(Positions)
	dec.Decode(MacdStates)
	dec.Decode(RsiHistory)
}

func RefreshAccountCache() {
	account, err := AlpacaClient.GetAccount()
	if err == nil {
		CachedBuyingPower, _ = strconv.ParseFloat(account.BuyingPower.String(), 64)
		DayTradeCount = int(account.DaytradeCount)
	}
}
