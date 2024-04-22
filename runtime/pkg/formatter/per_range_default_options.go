package formatter

func defaultGenericNumOptions() formatterRangeSpecsStrategy {
	return formatterRangeSpecsStrategy{
		rangeSpecs: []rangeFormatSpec{
			*newRangeFormatSpec(-2, 3, 3, 2, 0, false),
		},
		numKind:               numAny,
		defaultMaxDigitsRight: 1,
	}
}

func defaultPercentOptions() formatterRangeSpecsStrategy {
	return formatterRangeSpecsStrategy{
		rangeSpecs: []rangeFormatSpec{
			*newRangeFormatSpec(-2, 3, 3, 1, 0, false),
		},
		numKind:               numPercent,
		defaultMaxDigitsRight: 1,
	}
}

func defaultCurrencyOptions(numKind numberKind) formatterRangeSpecsStrategy {
	return formatterRangeSpecsStrategy{
		rangeSpecs: []rangeFormatSpec{
			*newRangeFormatSpec(-2, 3, 3, 2, 0, true),
		},
		numKind:               numKind,
		defaultMaxDigitsRight: 1,
	}
}
