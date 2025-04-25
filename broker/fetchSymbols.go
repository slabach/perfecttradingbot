package broker

import (
	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"log"
)

func FetchSymbols(alpacaClient *alpaca.Client, marketClient *marketdata.Client, useLive bool, ignoreSymbols map[string]bool) []string {
	assets, err := alpacaClient.GetAssets(alpaca.GetAssetsRequest{Status: string(alpaca.OptionStatusActive)})
	if err != nil {
		log.Fatalf("error fetching asset list: %v", err)
	}

	var result []string
	var symbolsToCheck []string
	for _, asset := range assets {
		if asset.Tradable {
			symbolsToCheck = append(symbolsToCheck, asset.Symbol)
		}
	}

	feed := "iex"
	if useLive {
		feed = "sip"
	}

	batchSize := 100
	for i := 0; i < len(symbolsToCheck); i += batchSize {
		end := i + batchSize
		if end > len(symbolsToCheck) {
			end = len(symbolsToCheck)
		}
		batch := symbolsToCheck[i:end]
		snapshots, err := marketClient.GetSnapshots(batch, marketdata.GetSnapshotRequest{
			Feed:     feed,
			Currency: "USD",
		})
		if err != nil {
			log.Printf("error fetching snapshots batch: %v", err)
			continue
		}

		for sym, snapshot := range snapshots {
			if snapshot == nil || snapshot.DailyBar == nil || snapshot.LatestTrade == nil || ignoreSymbols[sym] {
				continue
			}
			last := snapshot.LatestTrade.Price
			vol := snapshot.DailyBar.Volume

			if last >= 0.1 && last <= 20.0 && vol >= 100000 && snapshot.PrevDailyBar != nil {
				result = append(result, sym)
			}

		}
	}

	return result
}
