package druid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"text/template"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
)

// Constants for default values and property keys
const (
	InputPropertiesKey          = "inputProperties"
	DefaultTimeFormat           = "2006-01-02T15:04:05"
	DefaultIntervalFormat       = "2006-01-02T15:04:05"
	DefaultInitialLookBackPeriod = 60 * time.Minute
	CoordinatorURLKey           = "coordinatorURL"
	DataSourceNameKey           = "dataSourceName"

	PreviousExecutionTimeKey   = "previousExecutionTime"
	PreviousIntervalEndTimeKey = "previousIntervalEndTime"
	PathKey                    = "path"
	PatternKey                 = "pattern"
	GranularityKey             = "gran"
	PeriodBeforeKey            = "periodBefore"
	MaxWorkKey                 = "maxWork"
	QuietPeriodKey             = "quietPeriod"
	CatchupKey                 = "catchup"
	IndexJsonKey               = "indexJson"
)

// druidIndexExecutor is responsible for executing Druid index tasks.
type druidIndexExecutor struct {
	connection      *connection
	objectStore     drivers.ObjectStore
	executorOptions *drivers.ModelExecutorOptions
}

// Verify that druidIndexExecutor implements drivers.ModelExecutor.
var _ drivers.ModelExecutor = &druidIndexExecutor{}

// Execute implements drivers.ModelExecutor.
func (d *druidIndexExecutor) Execute(ctx context.Context) (*drivers.ModelResult, error) {
	dataSource := getStringWithDefault(d.executorOptions.OutputProperties, DataSourceNameKey, d.executorOptions.ModelName)

	previousResult, err := d.getPreviousResult()
	if err != nil {
		return nil, fmt.Errorf("error getting previous result: %w", err)
	}

	previousExecutionTime, err := getTimeFromProperty(previousResult.Properties, PreviousExecutionTimeKey, DefaultTimeFormat)
	if err != nil {
		return nil, fmt.Errorf("error getting previous execution time: %w", err)
	}

	previousIntervalEndTime, err := getTimeFromProperty(previousResult.Properties, PreviousIntervalEndTimeKey, DefaultIntervalFormat)
	if err != nil {
		return nil, fmt.Errorf("error getting previous interval end time: %w", err)
	}

	inputProperties, err := getMapFromProperty(d.executorOptions.InputProperties, InputPropertiesKey)
	if err != nil {
		return nil, fmt.Errorf("error getting input properties: %w", err)
	}

	path, err := getStringFromProperty(inputProperties, PathKey)
	if err != nil {
		return nil, fmt.Errorf("error getting path: %w", err)
	}

	pattern, err := getStringFromProperty(inputProperties, PatternKey)
	if err != nil {
		return nil, fmt.Errorf("error getting pattern: %w", err)
	}

	granularity, err := getDurationFromProperty(inputProperties, GranularityKey, 0)
	if err != nil {
		return nil, fmt.Errorf("error getting granularity: %w", err)
	}

	periodBefore, err := getDurationFromProperty(d.executorOptions.OutputProperties, PeriodBeforeKey, 0)
	if err != nil {
		return nil, fmt.Errorf("error getting period before: %w", err)
	}

	maxWork, err := getDurationFromProperty(d.executorOptions.OutputProperties, MaxWorkKey, 0)
	if err != nil {
		return nil, fmt.Errorf("error getting max work: %w", err)
	}

	quietPeriod, err := getDurationFromProperty(d.executorOptions.OutputProperties, QuietPeriodKey, 0)
	if err != nil {
		return nil, fmt.Errorf("error getting quiet period: %w", err)
	}

	catchup, err := getBoolFromProperty(d.executorOptions.OutputProperties, CatchupKey)
	if err != nil {
		return nil, fmt.Errorf("error getting catchup flag: %w", err)
	}

	specJson, err := getStringFromProperty(d.executorOptions.OutputProperties, IndexJsonKey)
	if err != nil {
		return nil, fmt.Errorf("error getting index JSON: %w", err)
	}

	startTime := calculateStartTime(previousExecutionTime, previousIntervalEndTime, periodBefore, catchup, quietPeriod)
	endTime := calculateEndTime(startTime, maxWork, quietPeriod)

	if endTime.Before(startTime) {
		return nil, fmt.Errorf("end time is before start time")
	}

	intervals := []string{fmt.Sprintf("%s/%s", startTime.Format(DefaultIntervalFormat), endTime.Format(DefaultIntervalFormat))}
	prefixes := generatePrefixes(path, pattern, granularity, startTime, endTime)

	specJson, err = renderIndexJson(specJson, intervals, prefixes)
	if err != nil {
		return nil, fmt.Errorf("failed to render index JSON: %w", err)
	}

	coordinatorURL, ok := d.executorOptions.OutputProperties[CoordinatorURLKey].(string)
	if !ok {
		return nil, fmt.Errorf("coordinator URL is not a string or is missing")
	}

	dataSourceName, ok := d.executorOptions.OutputProperties[DataSourceNameKey].(string)
	if !ok {
		return nil, fmt.Errorf("data source name is not a string or is missing")
	}

	err = Ingest(coordinatorURL, specJson, dataSourceName, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest data: %w", err)
	}

	return &drivers.ModelResult{
		Connector: d.executorOptions.OutputConnector,
		Properties: map[string]interface{}{
			"dataSource":    fmt.Sprintf("%s-%s", d.executorOptions.ModelName, dataSource),
			PreviousExecutionTimeKey: time.Now().Format(DefaultTimeFormat),
			PreviousIntervalEndTimeKey: endTime.Format(DefaultTimeFormat),
		},
	}, nil
}

// getPreviousResult retrieves the previous result or initializes it if not available.
func (d *druidIndexExecutor) getPreviousResult() (*drivers.ModelResult, error) {
	if d.executorOptions.PreviousResult == nil {
		initialLookBackPeriod, err := getDurationFromProperty(d.executorOptions.OutputProperties, "initialLookBackPeriod", DefaultInitialLookBackPeriod)
		if err != nil {
			return nil, fmt.Errorf("error getting initial lookback period: %w", err)
		}
		return &drivers.ModelResult{
			Properties: map[string]interface{}{
				PreviousExecutionTimeKey:   time.Now().Add(-initialLookBackPeriod).Format(DefaultTimeFormat),
				PreviousIntervalEndTimeKey: time.Now().Add(-initialLookBackPeriod).Format(DefaultTimeFormat),
			},
		}, nil
	}
	return d.executorOptions.PreviousResult, nil
}

// generatePrefixes generates a list of prefixes based on the provided path, pattern, and granularity.
func generatePrefixes(path, pattern string, gran time.Duration, startTime, endTime time.Time) []string {
	var prefixes []string

	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	for t := startTime.Add(gran); t.Before(endTime); t = t.Add(gran) {
		prefix := fmt.Sprintf("%s/%s", path, t.Format(pattern))
		prefixes = append(prefixes, prefix)
	}

	return prefixes
}

// renderIndexJson renders the index JSON template with the provided intervals and prefixes.
func renderIndexJson(indexJson string, intervals, prefixes []string) (string, error) {
	t := template.Must(template.New("indexJson").Funcs(template.FuncMap{
		"toJson": func(v interface{}) (string, error) {
			a, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(a), nil
		},
	}).Parse(indexJson))

	var buf bytes.Buffer
	err := t.Execute(&buf, map[string]interface{}{
		"intervals": intervals,
		"prefixes":  prefixes,
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute index template: %w", err)
	}

	return buf.String(), nil
}

// getDurationFromProperty retrieves a duration from properties or returns the default value.
func getDurationFromProperty(props map[string]interface{}, key string, defaultValue time.Duration) (time.Duration, error) {
	if value, ok := props[key]; ok {
		if durationStr, ok := value.(string); ok {
			return time.ParseDuration(durationStr)
		}
		return 0, fmt.Errorf("invalid duration format for %s", key)
	}
	return defaultValue, nil
}

// getMapFromProperty retrieves a map from properties.
func getMapFromProperty(props map[string]interface{}, key string) (map[string]interface{}, error) {
	if value, ok := props[key]; ok {
		if m, ok := value.(map[string]interface{}); ok {
			return m, nil
		}
		return nil, fmt.Errorf("%s is expected to be a map", key)
	}
	return nil, fmt.Errorf("%s is expected", key)
}

// getStringFromProperty retrieves a string from properties.
func getStringFromProperty(props map[string]interface{}, key string) (string, error) {
	if value, ok := props[key]; ok {
		if str, ok := value.(string); ok {
			return str, nil
		}
		return "", fmt.Errorf("%s is expected to be a string", key)
	}
	return "", fmt.Errorf("%s is expected", key)
}

// getBoolFromProperty retrieves a boolean from properties.
func getBoolFromProperty(props map[string]interface{}, key string) (bool, error) {
	value, ok := props[key]
	if !ok {
		return false, fmt.Errorf("%s is expected", key)
	}

	switch v := value.(type) {
	case string:
		return strconv.ParseBool(v)
	case bool:
		return v, nil
	default:
		return false, fmt.Errorf("invalid boolean format for %s", key)
	}
}

// getTimeFromProperty retrieves a time from properties using the provided format.
func getTimeFromProperty(props map[string]interface{}, key, format string) (time.Time, error) {
	if value, ok := props[key]; ok {
		return time.Parse(format, value.(string))
	}
	return time.Time{}, fmt.Errorf("%s is expected", key)
}

// calculateStartTime calculates the start time based on the provided parameters.
func calculateStartTime(previousExecutionTime, previousIntervalEndTime time.Time, periodBefore time.Duration, catchup bool, quietPeriod time.Duration) time.Time {
	if catchup {
		return previousIntervalEndTime.Add(-periodBefore)
	}
	return previousExecutionTime.Add(-quietPeriod).Add(-periodBefore)
}

// calculateEndTime calculates the end time based on the start time and provided durations.
func calculateEndTime(startTime time.Time, maxWork, quietPeriod time.Duration) time.Time {
	maxEndTime := time.Now().Add(-quietPeriod)
	if maxEndTime.Before(startTime.Add(maxWork)) {
		return maxEndTime
	}
	return startTime.Add(maxWork)
}

// getStringWithDefault retrieves a string from properties or returns the default value.
func getStringWithDefault(props map[string]interface{}, key, defaultValue string) string {
	if value, ok := props[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}
