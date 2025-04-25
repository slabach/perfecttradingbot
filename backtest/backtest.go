package backtest

import (
	"log"
	"os"
	"perfectTradingBot/global"
	"perfectTradingBot/types"
	"strings"
	"time"
)

func RunBacktestMode() {
	symbolList := os.Getenv("BACKTEST_SYMBOLS")
	if symbolList == "" {
		log.Fatal("BACKTEST_SYMBOLS is required")
	}
	symbols := parseSymbols(symbolList)

	start, err := time.Parse(time.RFC3339, os.Getenv("BACKTEST_START"))
	if err != nil {
		log.Fatalf("Invalid BACKTEST_START format: %v", err)
	}
	end, err := time.Parse(time.RFC3339, os.Getenv("BACKTEST_END"))
	if err != nil {
		log.Fatalf("Invalid BACKTEST_END format: %v", err)
	}

	for _, sym := range symbols {
		RunBacktest(sym, start, end)
	}
}

func parseSymbols(list string) []string {
	var result []string
	for _, sym := range strings.Split(list, ",") {
		if s := strings.TrimSpace(sym); s != "" {
			result = append(result, s)
		}
	}
	return result
}

func RunBacktest(symbol string, start, end time.Time) {
	log.Printf("Starting backtest for %s from %s to %s", symbol, start.Format(time.RFC3339), end.Format(time.RFC3339))

	//req := marketdata.GetBarsRequest{
	//	TimeFrame: marketdata.OneMin,
	//	Start:     start,
	//	End:       end,
	//	Feed:      "iex",
	//}
	//bars, err := global.MarketClient.GetBars(symbol, req)
	//if err != nil {
	//	log.Fatalf("Failed to fetch historical bars for %s: %v", symbol, err)
	//}

	// Reset state
	global.CashBalance = global.InitialBalance
	global.TotalProfitLoss = 0
	global.Positions = make(map[string]types.Position)
	global.MacdStates = map[string]*types.MACDState{symbol: {}}
	global.RsiHistory = map[string][]float64{}
	global.VolHistory = map[string][]float64{}
	global.BarBuffer = map[string]*types.AggregatedBar{}
	global.ReclaimConfirmed = map[string]bool{}
	global.PullbackTracked = map[string]bool{}
	global.BreakoutHighs = map[string]float64{}
	global.LastTradeTime = map[string]time.Time{}
	global.AnticipatoryFlags = map[string]bool{}
	global.VwapData = map[string]*types.VWAPState{}

	//for _, bar := range bars {
	//	//timestamp := bar.Timestamp.In(time.FixedZone("EST", -5*60*60))
	//	//handleBar(stream.Bar{
	//	//	Symbol:    symbol,
	//	//	Open:      bar.Open,
	//	//	High:      bar.High,
	//	//	Low:       bar.Low,
	//	//	Close:     bar.Close,
	//	//	Volume:    bar.Volume,
	//	//	Timestamp: timestamp,
	//	//})
	//}

	log.Printf("Backtest complete for %s. Final Balance: $%.2f | Net P&L: $%.2f\n", symbol, global.CashBalance, global.TotalProfitLoss)
}
