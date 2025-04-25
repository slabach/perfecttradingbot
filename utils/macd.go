package utils

import (
	"perfectTradingBot/types"
	"sync"
)

var (
	alphaFast   = 2.0 / float64(fastPeriod+1)
	alphaSlow   = 2.0 / float64(slowPeriod+1)
	alphaSignal = 2.0 / float64(signalPeriod+1)
)

const (
	fastPeriod   = 8
	slowPeriod   = 17
	signalPeriod = 9
)

//func updateMACD(st *types.MACDState, price float64) {
//	if !st.Initialized {
//		st.FastEMA = price
//		st.SlowEMA = price
//		st.SignalEMA = 0.0
//		st.LastMACD = 0.0
//		st.LastSignal = 0.0
//		st.Initialized = true
//		st.LastCrossUp = false
//		return
//	}
//	st.FastEMA = alphaFast*price + (1-alphaFast)*st.FastEMA
//	st.SlowEMA = alphaSlow*price + (1-alphaSlow)*st.SlowEMA
//	currentMACD := st.FastEMA - st.SlowEMA
//	st.SignalEMA = alphaSignal*currentMACD + (1-alphaSignal)*st.SignalEMA
//	st.LastMACD = currentMACD
//	st.LastSignal = st.SignalEMA
//}

func updateMACDWorker(jobs <-chan types.MACDJob, results chan<- types.MACDResult, states map[string]*types.MACDState, lock *sync.RWMutex, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		lock.RLock()
		st, ok := states[job.Symbol]
		lock.RUnlock()
		if !ok {
			st = &types.MACDState{}
		}

		if !st.Initialized {
			st.FastEMA = job.Price
			st.SlowEMA = job.Price
			st.SignalEMA = 0.0
			st.LastMACD = 0.0
			st.LastSignal = 0.0
			st.Initialized = true
			st.LastCrossUp = false
		} else {
			st.FastEMA = alphaFast*job.Price + (1-alphaFast)*st.FastEMA
			st.SlowEMA = alphaSlow*job.Price + (1-alphaSlow)*st.SlowEMA
			macd := st.FastEMA - st.SlowEMA
			st.SignalEMA = alphaSignal*macd + (1-alphaSignal)*st.SignalEMA
			st.LastMACD = macd
			st.LastSignal = st.SignalEMA
		}

		results <- types.MACDResult{Symbol: job.Symbol, State: st}
	}
}

func RunMACDPool(prices map[string]float64, states map[string]*types.MACDState, lock *sync.RWMutex) map[string]*types.MACDState {
	jobs := make(chan types.MACDJob, len(prices))
	results := make(chan types.MACDResult, len(prices))
	var wg sync.WaitGroup

	workers := 8
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go updateMACDWorker(jobs, results, states, lock, &wg)
	}

	for sym, price := range prices {
		jobs <- types.MACDJob{Symbol: sym, Price: price}
	}
	close(jobs)

	wg.Wait()
	close(results)

	updatedStates := make(map[string]*types.MACDState)
	for res := range results {
		updatedStates[res.Symbol] = res.State
		lock.Lock()
		states[res.Symbol] = res.State
		lock.Unlock()
	}

	return updatedStates
}
