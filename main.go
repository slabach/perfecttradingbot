// main.go - Full restored app with trend filtering implementation

package main

import (
	"context"
	"fmt"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"log"
	"os"
	"os/signal"
	"perfectTradingBot/backtest"
	"perfectTradingBot/broker"
	"perfectTradingBot/global"
	"perfectTradingBot/types"
	"strconv"
	"syscall"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	_ "github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	global.LoadStateFromDisk()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}
	initConfig()
	initAlpacaClient()

	if os.Getenv("BACKTEST_MODE") == "true" {
		backtest.RunBacktestMode()
		return
	}

	global.RefreshAccountCache()
	go func() {
		for {
			global.RefreshAccountCache()
			time.Sleep(1 * time.Minute)
		}
	}()

	account, err := global.AlpacaClient.GetAccount()
	if err != nil {
		log.Printf("error fetching account: %v", err)
	}
	global.InitialBalance, _ = strconv.ParseFloat(account.BuyingPower.String(), 64)

	global.Symbols = broker.FetchSymbols(global.AlpacaClient, global.MarketClient, global.UseLive, global.IgnoreSymbols)
	fmt.Println(global.Symbols)
	if len(global.Symbols) == 0 {
		log.Fatal("No symbols found for today's trading criteria")
	}
	for _, sym := range global.Symbols {
		global.MacdStates[sym] = &types.MACDState{}
	}

	ctx := context.Background()
	feed := "iex"
	if global.UseLive {
		feed = "sip"
	}
	ws := stream.NewStocksClient(feed)

	err = ws.Connect(ctx)
	if err != nil {
		log.Fatalf("failed to connect to Alpaca stream: %v", err)
	}

	err = ws.SubscribeToBars(func(bar stream.Bar) {
		broker.HandleBar(bar)
	}, global.Symbols...)
	if err != nil {
		log.Fatalf("failed to subscribe to bars: %v", err)
	}

	//err = ws.SubscribeToTrades(func(trade stream.Trade) {
	//	broker.HandleTrade(trade)
	//}, global.Symbols...)
	//if err != nil {
	//	log.Fatalf("failed to subscribe to bars: %v", err)
	//}

	log.Printf("Subscribed to: %v\n", global.Symbols)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Println("Shutting down gracefully...")
		global.StateLock.Lock()
		global.SaveStateToDisk()
		global.StateLock.Unlock()
		os.Exit(0)
	}()

	go global.PersistStatePeriodically()
	select {}
}

func initConfig() {
	if v := os.Getenv("TRADE_FUND_PERCENT"); v != "" {
		if p, err := strconv.ParseFloat(v, 64); err == nil {
			global.TradeFundPercent = p
		}
	}
	if os.Getenv("LIVE_TRADING") == "true" {
		global.UseLive = true
	}

	global.MinHist, _ = strconv.ParseFloat(os.Getenv("MIN_HISTOGRAM"), 64)
	global.MinSlope, _ = strconv.ParseFloat(os.Getenv("MACD_SLOPE_THRESHOLD"), 64)
	global.LongEMAPeriod, _ = strconv.Atoi(os.Getenv("LONG_EMA_PERIOD"))
	global.RsiMin, _ = strconv.ParseFloat(os.Getenv("RSI_MIN"), 64)
	global.RsiMax, _ = strconv.ParseFloat(os.Getenv("RSI_MAX"), 64)
	global.RequireTrend = os.Getenv("REQUIRE_TREND") == "true"
	global.UseRSIFilter = os.Getenv("USE_RSI_FILTER") == "true"
	//global.UseLive = os.Getenv("LIVE_TRADING") == "false"
}

func initAlpacaClient() {
	baseURL := "https://paper-api.alpaca.markets"
	if global.UseLive {
		baseURL = "https://api.alpaca.markets"
	}
	global.AlpacaClient = alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    os.Getenv("APCA_API_KEY_ID"),
		APISecret: os.Getenv("APCA_API_SECRET_KEY"),
		BaseURL:   baseURL,
	})

	global.MarketClient = marketdata.NewClient(marketdata.ClientOpts{
		APIKey:    os.Getenv("APCA_API_KEY_ID"),
		APISecret: os.Getenv("APCA_API_SECRET_KEY"),
	})

	log.Printf("Trading mode: %s\n", map[bool]string{true: "LIVE", false: "SIMULATION"}[global.UseLive])
}
