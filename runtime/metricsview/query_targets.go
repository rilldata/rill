package metricsview

import "encoding/json"

type TargetsQuery struct {
	MetricsView string     `json:"metrics_view" mapstructure:"metrics_view"`
	Measures    []string   `json:"measures" mapstructure:"measures"`
	Target      string     `json:"target" mapstructure:"target"` // Target identifier (name)
	TimeRange   *TimeRange `json:"time_range" mapstructure:"time_range"`
	TimeZone    string     `json:"time_zone" mapstructure:"time_zone"`
	TimeGrain   TimeGrain  `json:"time_grain" mapstructure:"time_grain"` // Time grain from query (for filtering by grain column)
	Priority    int        `json:"priority" mapstructure:"priority"`
}

func (q *TargetsQuery) AsMap() (map[string]any, error) {
	// We do a JSON roundtrip to convert to a map[string]any.
	// We don't use mapstructure here because it doesn't correctly handle time.Time roundtrips to a map[string]any, even with decoder hooks.
	// And anyway, since JSON is the usual entrypoint for TimeRange maps, this is more representative of real usage.
	data, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	var res map[string]any
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

