package metricsview

import "github.com/mitchellh/mapstructure"

type AnnotationsQuery struct {
	MetricsView string     `mapstructure:"metrics_view"`
	Measures    []string   `mapstructure:"measures"`
	TimeRange   *TimeRange `mapstructure:"time_range"`
	Limit       *int64     `mapstructure:"limit"`
	Offset      *int64     `mapstructure:"offset"`
	TimeZone    string     `mapstructure:"time_zone"`
	TimeGrain   TimeGrain  `mapstructure:"time_grain"`
	Priority    int        `mapstructure:"priority"`
}

func (q *AnnotationsQuery) AsMap() (map[string]any, error) {
	queryMap := make(map[string]any)
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     &queryMap,
		DecodeHook: timeDecodeFunc,
	})
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(q)
	if err != nil {
		return nil, err
	}
	return queryMap, nil
}
