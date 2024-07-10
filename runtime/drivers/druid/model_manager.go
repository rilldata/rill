package druid

import (
	"context"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

type ModelInputProperties struct {
	Path          string `mapstructure:"path"`
	Pattern       string `mapstructure:"pattern"`
	Granularity   string `mapstructure:"gran"`
	Format        string `mapstructure:"format"`
	FilePattern   string `mapstructure:"file_pattern"`
	RetriesPeriod string `mapstructure:"retry_period"`
	MaxRetries    int    `mapstructure:"max_retries"`
}

func (p *ModelInputProperties) Validate() error {
	if p.Path == "" {
		return fmt.Errorf("missing property 'path'")
	}
	if p.Pattern == "" {
		return fmt.Errorf("missing property 'pattern'")
	}
	if p.Granularity == "" {
		return fmt.Errorf("missing property 'gran'")
	}

	_, err := time.ParseDuration(p.Granularity)
	if err != nil {
		return fmt.Errorf("invalid value for property 'gran': %w", err)
	}
	return nil
}

type ModelOutputProperties struct {
	Connector             string `mapstructure:"connector"`
	DataSource            string `mapstructure:"datasource"`
	InitialLookBackPeriod string `mapstructure:"initial_look_back_period"`
	PeriodBefore          string `mapstructure:"period_before"`
	QuietPeriod           string `mapstructure:"quiet_period"`
	Catchup               bool   `mapstructure:"catchup"`
	MaxWork               string `mapstructure:"max_work"`
	CoordinatorURL        string `mapstructure:"coordinator_url"`
	DataSourceName        string `mapstructure:"datasource_name"`
	SpecJson              string `mapstructure:"spec_json"`
	TimeoutPeriod         string `mapstructure:"timeout_period"`
}

func (p *ModelOutputProperties) Validate(opts *drivers.ModelExecutorOptions) error {
	if p.DataSource == "" {
		return fmt.Errorf("missing property 'datasource'")
	}
	if p.CoordinatorURL == "" {
		return fmt.Errorf("missing property 'coordinatorURL'")
	}
	if p.DataSourceName == "" {
		return fmt.Errorf("missing property 'dataSourceName'")
	}
	if p.SpecJson == "" {
		return fmt.Errorf("missing property 'specJson'")
	}
	return nil

}

type ModelResultProperties struct {
	DataSource              string `mapstructure:"datasource"`
	PreviousExecutionTime   string `mapstructure:"previous_execution_time"`
	PreviousIntervalEndTime string `mapstructure:"previous_interval_end_time"`
}

func (c *connection) Rename(ctx context.Context, res *drivers.ModelResult, newName string, env *drivers.ModelEnv) (*drivers.ModelResult, error) {

	resPropsMap := map[string]interface{}{}
	if err := mapstructure.WeakDecode(res.Properties, &resPropsMap); err != nil {
		return nil, fmt.Errorf("failed to parse previous result properties: %w", err)
	}

	return &drivers.ModelResult{
		Connector:  res.Connector,
		Properties: resPropsMap,
		Table:      res.Table,
	}, nil
}

func (c *connection) Exists(ctx context.Context, res *drivers.ModelResult) (bool, error) {
	olap, ok := c.AsOLAP(c.instanceID)
	if !ok {
		return false, fmt.Errorf("connector is not an OLAP")
	}

	_, err := olap.InformationSchema().Lookup(ctx, "", "", res.Table)
	return err == nil, nil

}

func (c *connection) Delete(ctx context.Context, res *drivers.ModelResult) error {
	// do nothing
	return nil
}
