package marketdata

import (
	"encoding/csv"
	"fmt"
	"os"
	"perfectTradingBot/types"
	"strconv"
	"time"
)

func FetchHistoricalBars(contractID string) ([]types.Bar, error) {
	filename := fmt.Sprintf("data/%s_bars.csv", contractID)
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open historical bar file: %w", err)
	}
	defer file.Close()

	rdr := csv.NewReader(file)
	records, err := rdr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not read CSV: %w", err)
	}

	var bars []types.Bar
	for i, row := range records {
		if i == 0 {
			continue // skip header
		}
		t, _ := time.Parse("2025-01-02 15:04:05", row[0])
		op, _ := strconv.ParseFloat(row[1], 64)
		high, _ := strconv.ParseFloat(row[2], 64)
		low, _ := strconv.ParseFloat(row[3], 64)
		close, _ := strconv.ParseFloat(row[4], 64)
		vol, _ := strconv.Atoi(row[5])

		bars = append(bars, types.Bar{
			Timestamp: t,
			Open:      op,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    vol,
		})
	}
	return bars, nil
}

func NewBarBacktestProvider(bars []types.Bar) <-chan types.TickData {
	tickChan := make(chan types.TickData, len(bars))
	go func() {
		defer close(tickChan)
		for _, bar := range bars {
			tick := types.TickData{
				Timestamp: bar.Timestamp.Unix(),
				Last:      bar.Close,
				Volume:    float64(bar.Volume),
			}
			tickChan <- tick
		}
	}()
	return tickChan
}
