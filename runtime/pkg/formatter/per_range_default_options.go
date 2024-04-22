package formatter

func defaultNoneOptions() FormatterOptionsNoneStrategy {
	return FormatterOptionsNoneStrategy{
		NumberKind: ANY,
	}
}

func defaultGenericNumOptions() FormatterRangeSpecsStrategy {
	return FormatterRangeSpecsStrategy{
		formatterOptionsCommon: formatterOptionsCommon{
			NumberKind: ANY,
		},
		RangeSpecs: []rangeFormatSpec{
			*newRangeFormatSpec(-2, 3, 3, 2, 0, false),
		},
		DefaultMaxDigitsRight: 1,
	}
}

func defaultPercentOptions() FormatterRangeSpecsStrategy {
	return FormatterRangeSpecsStrategy{
		formatterOptionsCommon: formatterOptionsCommon{
			NumberKind: PERCENT,
		},
		RangeSpecs: []rangeFormatSpec{
			*newRangeFormatSpec(-2, 3, 3, 1, 0, false),
		},
		DefaultMaxDigitsRight: 1,
	}
}

func defaultCurrencyOptions(numberKind numberKind) FormatterRangeSpecsStrategy {
	return FormatterRangeSpecsStrategy{
		formatterOptionsCommon: formatterOptionsCommon{
			NumberKind: numberKind,
		},
		RangeSpecs: []rangeFormatSpec{
			*newRangeFormatSpec(-2, 3, 3, 2, 0, true),
		},
		DefaultMaxDigitsRight: 1,
	}
}
