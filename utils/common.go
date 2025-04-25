package utils

import (
	"fmt"
	"log"
	"math"
	"os"
	"perfectTradingBot/global"
	"time"
)

func CalculateOrderSize(price float64) int {
	maxSpend := global.CachedBuyingPower * global.TradeFundPercent
	shares := int(math.Floor(maxSpend / price))
	if shares < 1 {
		return 0
	}
	return shares
}

func LogTradeCSV(sym string, entry float64, exit float64, shares int, pnl float64, entryTime time.Time) {
	fileDir := os.Getenv("FILE_DIR")
	if fileDir == "" {
		fileDir = "."
	}

	f, err := os.OpenFile(fmt.Sprintf("%s/trades.csv", fileDir), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("error logging trade to CSV: %v", err)
		return
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Printf("error logging trade to CSV: %v", err)
		}
	}(f)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	duration := time.Since(entryTime).Round(time.Second).String()
	line := fmt.Sprintf("%s,%s,%d,%.2f,%.2f,%.2f,%s\n", timestamp, sym, shares, entry, exit, pnl, duration)
	if _, err := f.WriteString(line); err != nil {
		log.Printf("error writing trade log line: %v", err)
	}
}
