package druid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

// Constants for default values and property keys
const (
	DefaultTimeFormat            = time.RFC3339
	DefaultIntervalFormat        = time.RFC3339
	DefaultInitialLookBackPeriod = 60 * time.Minute
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
func (e *druidIndexExecutor) Execute(ctx context.Context) (*drivers.ModelResult, error) {

	inputProperties := &ModelInputProperties{}
	err := mapstructure.WeakDecode(e.executorOptions.InputProperties, inputProperties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}

	outputProperties := &ModelOutputProperties{}
	err = mapstructure.WeakDecode(e.executorOptions.OutputProperties, outputProperties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}

	var initialLookBackPeriod time.Duration
	if outputProperties.InitialLookBackPeriod == "" {
		initialLookBackPeriod = DefaultInitialLookBackPeriod
	} else {
		initialLookBackPeriod, err = time.ParseDuration(outputProperties.InitialLookBackPeriod)
		if err != nil {
			return nil, fmt.Errorf("failed to parse initial look back period: %w", err)
		}
	}

	previousResult, err := e.getPreviousResult(initialLookBackPeriod, DefaultTimeFormat)
	if err != nil {
		return nil, fmt.Errorf("error getting previous result: %w", err)
	}

	prevModelResultProperties := &ModelResultProperties{}
	err = mapstructure.WeakDecode(previousResult.Properties, prevModelResultProperties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse previous result properties: %w", err)
	}

	previousExecutionTime, err := time.Parse(time.RFC3339, prevModelResultProperties.PreviousExecutionTime)
	if err != nil {
		return nil, fmt.Errorf("error getting previous execution time: %w", err)
	}

	previousIntervalEndTime, err := time.Parse(time.RFC3339, prevModelResultProperties.PreviousIntervalEndTime)
	if err != nil {
		return nil, fmt.Errorf("error getting previous interval end time: %w", err)
	}

	periodBefore, err := time.ParseDuration(outputProperties.PeriodBefore)
	if err != nil {
		return nil, fmt.Errorf("failed to parse period before: %w", err)
	}

	maxWork, err := time.ParseDuration(outputProperties.MaxWork)
	if err != nil {
		return nil, fmt.Errorf("failed to parse max work: %w", err)
	}

	quietPeriod, err := time.ParseDuration(outputProperties.QuietPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to parse quiet period: %w", err)
	}

	startTime := calculateStartTime(previousExecutionTime, previousIntervalEndTime, periodBefore, outputProperties.Catchup, quietPeriod)
	endTime := calculateEndTime(startTime, maxWork, quietPeriod)

	if endTime.Before(startTime) {
		return nil, fmt.Errorf("end time is before start time")
	}

	gran, err := time.ParseDuration(inputProperties.Granularity)
	if err != nil {
		return nil, fmt.Errorf("failed to parse granularity: %w", err)
	}

	intervals := []string{fmt.Sprintf("%s/%s", startTime.Format(DefaultIntervalFormat), endTime.Format(DefaultIntervalFormat))}
	prefixes := generatePrefixes(inputProperties.Path, inputProperties.Pattern, gran, startTime, endTime)

	specJson, err := renderIndexJson(outputProperties.SpecJson, intervals, prefixes)
	if err != nil {
		return nil, fmt.Errorf("failed to render index JSON: %w", err)
	}

	var timeout time.Duration
	if outputProperties.TimeoutPeriod == "" {
		timeout = 10 * time.Minute
	} else {
		timeout, err = time.ParseDuration(outputProperties.TimeoutPeriod)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timeout period: %w", err)
		}
	}

	err = Ingest(outputProperties.CoordinatorURL, specJson, outputProperties.DataSourceName, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest data: %w", err)
	}

	var datasourceName string
	if outputProperties.DataSourceName == "" {
		datasourceName = e.executorOptions.ModelName
	} else {
		datasourceName = outputProperties.DataSourceName
	}

	modelResultProperties := &ModelResultProperties{
		DataSource:              datasourceName,
		PreviousExecutionTime:   time.Now().Format(DefaultTimeFormat),
		PreviousIntervalEndTime: endTime.Format(DefaultTimeFormat),
	}
	modelResultPropertiesMap := map[string]interface{}{}
	err = mapstructure.WeakDecode(modelResultProperties, &modelResultPropertiesMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}

	return &drivers.ModelResult{
		Connector:  e.executorOptions.OutputConnector,
		Properties: modelResultPropertiesMap,
		Table:      datasourceName,
	}, nil
}

// getPreviousResult retrieves the previous result or initializes it if not available.
func (e *druidIndexExecutor) getPreviousResult(initialLookBackPeriod time.Duration, DefaultTimeFormat string) (*drivers.ModelResult, error) {
	if e.executorOptions.PreviousResult == nil {
		modelProperties := &ModelResultProperties{
			PreviousExecutionTime:   time.Now().Add(-initialLookBackPeriod).Format(DefaultTimeFormat),
			PreviousIntervalEndTime: time.Now().Add(-initialLookBackPeriod).Format(DefaultTimeFormat),
		}
		modelPropertiesMap := map[string]interface{}{}
		err := mapstructure.WeakDecode(modelProperties, &modelPropertiesMap)
		if err != nil {
			return nil, fmt.Errorf("failed to encode result properties: %w", err)
		}
		return &drivers.ModelResult{Properties: modelPropertiesMap}, nil
	}
	return e.executorOptions.PreviousResult, nil
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
