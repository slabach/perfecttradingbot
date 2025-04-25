package utils

import (
	"perfectTradingBot/global"
	"perfectTradingBot/types"
)

func UpdateVWAP(sym string, price float64, volume float64) float64 {
	v := global.VwapData[sym]
	if v == nil {
		v = &types.VWAPState{}
		global.VwapData[sym] = v
	}
	v.TotalPV += price * volume
	v.TotalVol += volume
	if v.TotalVol == 0 {
		return price
	}
	return v.TotalPV / v.TotalVol
}
