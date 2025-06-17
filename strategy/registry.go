package strategy

var activeStrategies = []Strategy{
	&BreakoutStrategy{},
	// &MeanReversionStrategy{},
	// &OpeningRangeBreakout{},
}

func GetStrategies() []Strategy {
	return activeStrategies
}
