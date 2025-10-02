package metricsview

type SearchQuery struct {
	MetricsView string      `mapstructure:"metrics_view"`
	Dimensions  []string    `mapstructure:"dimensions"`
	Search      string      `mapstructure:"search"`
	Where       *Expression `mapstructure:"where"`
	Having      *Expression `mapstructure:"having"`
	TimeRange   *TimeRange  `mapstructure:"time_range"`
	Limit       *int64      `mapstructure:"limit"`
}

type SearchResult struct {
	Dimension string
	Value     any
}
