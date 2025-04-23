package parser

import (
	"fmt"
	"strconv"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/robfig/cron/v3"
)

// ScheduleYAML is the raw structure of a refresh schedule clause defined in YAML.
// This does not represent a stand-alone YAML file, just a partial used in other structs.
type ScheduleYAML struct {
	RefUpdate *bool  `yaml:"ref_update" mapstructure:"ref_update"`
	Cron      string `yaml:"cron" mapstructure:"cron"`
	Every     string `yaml:"every" mapstructure:"every"`
	TimeZone  string `yaml:"time_zone" mapstructure:"time_zone"`
	Disable   bool   `yaml:"disable" mapstructure:"disable"`
	RunInDev  bool   `yaml:"run_in_dev" mapstructure:"run_in_dev"`
}

func (p *Parser) parseScheduleYAML(raw *ScheduleYAML) (*runtimev1.Schedule, error) {
	s := &runtimev1.Schedule{
		RefUpdate: true, // By default, refresh on updates to refs
	}

	if raw == nil {
		return s, nil
	}

	if raw.Disable {
		s.RefUpdate = false
		s.Disable = true
		return s, nil
	}

	// Enforce run_in_dev only for scheduled refreshes. We always honor ref_update even in dev.
	skipScheduledRefresh := !raw.RunInDev && p.isDev()

	if raw.RefUpdate != nil {
		s.RefUpdate = *raw.RefUpdate
	}

	if !skipScheduledRefresh && raw.Cron != "" {
		_, err := cron.ParseStandard(raw.Cron)
		if err != nil {
			return nil, fmt.Errorf("invalid cron schedule: %w", err)
		}
		s.Cron = raw.Cron
	}

	if !skipScheduledRefresh && raw.Every != "" {
		d, err := parseDuration(raw.Every)
		if err != nil {
			return nil, fmt.Errorf("invalid ticker: %w", err)
		}
		s.TickerSeconds = uint32(d.Seconds())
	}

	if raw.TimeZone != "" {
		_, err := time.LoadLocation(raw.TimeZone)
		if err != nil {
			return nil, fmt.Errorf("invalid time zone: %w", err)
		}
		s.TimeZone = raw.TimeZone
	}

	return s, nil
}

// parseDuration parses a value into a time duration.
// If no unit is specified, it assumes the value is in seconds.
func parseDuration(v any) (time.Duration, error) {
	switch v := v.(type) {
	case int:
		return time.Duration(v) * time.Second, nil
	case string:
		// Try parsing as an int first
		res, err := strconv.Atoi(v)
		if err == nil {
			return time.Duration(res) * time.Second, nil
		}
		// Try parsing as a Go duration string
		d, err := time.ParseDuration(v)
		if err == nil {
			return d, nil
		}
		// Try parsing as an ISO 8601 duration string
		id, err := duration.ParseISO8601(v)
		if err == nil {
			d, ok := id.EstimateNative()
			if !ok {
				return 0, fmt.Errorf("time duration string %q can't be resolved to an absolute duration", v)
			}
			return d, nil
		}
		// Give up
		return 0, fmt.Errorf("invalid time duration string %q", v)
	default:
		return 0, fmt.Errorf("invalid time duration value <%v>", v)
	}
}
