package formatter

var orderOfMagTextToShortScaleSuffix = map[string]string{
	"E0":  "",
	"E3":  "k",
	"E6":  "M",
	"E9":  "B",
	"E12": "T",
	"E15": "Q",
}

// Converts a suffix like "E3" to "k", or returns the input if no conversion is available.
func shortScaleSuffixIfAvailableForStr(suffixIn string) string {
	if suffix, ok := orderOfMagTextToShortScaleSuffix[suffixIn]; ok {
		return suffix
	}
	return suffixIn
}
