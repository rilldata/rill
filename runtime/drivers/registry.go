package drivers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// RegistryStore is implemented by drivers capable of storing and looking up instances and repos.
type RegistryStore interface {
	FindInstances(ctx context.Context) ([]*Instance, error)
	FindInstance(ctx context.Context, id string) (*Instance, error)
	CreateInstance(ctx context.Context, instance *Instance) error
	DeleteInstance(ctx context.Context, id string) error
	EditInstance(ctx context.Context, instance *Instance) error
}

// Instance represents a single data project, meaning one OLAP connection, one repo connection,
// and one catalog connection.
type Instance struct {
	// Identifier
	ID string
	// Environment is the environment that the instance represents
	Environment string
	// ProjectDisplayName is the display name from rill.yaml
	ProjectDisplayName string `db:"project_display_name"`
	// Driver name to connect to for OLAP
	OLAPConnector string
	// ProjectOLAPConnector is an override of OLAPConnector that may be set in rill.yaml.
	// NOTE: Hopefully we can merge this with OLAPConnector if/when we remove the ability to set OLAPConnector using flags.
	ProjectOLAPConnector string
	// Driver name for reading/editing code artifacts
	RepoConnector string
	// Driver name for the admin service managing the deployment (optional)
	AdminConnector string
	// Driver name for the AI service (optional)
	AIConnector string
	// ProjectAIConnector is an override of AIConnector that may be set in rill.yaml.
	ProjectAIConnector string
	// Driver name for catalog
	CatalogConnector string
	// CreatedOn is when the instance was created
	CreatedOn time.Time `db:"created_on"`
	// UpdatedOn is when the instance was last updated in the registry
	UpdatedOn time.Time `db:"updated_on"`
	// Instance specific connectors
	Connectors []*runtimev1.Connector `db:"connectors"`
	// ProjectConnectors contains default connectors from rill.yaml
	ProjectConnectors []*runtimev1.Connector `db:"project_connectors"`
	// Variables contains user-provided variables
	Variables map[string]string `db:"variables"`
	// ProjectVariables contains default variables from rill.yaml
	// (NOTE: This can always be reproduced from rill.yaml, so it's really just a handy cache of the values.)
	ProjectVariables map[string]string `db:"project_variables"`
	// FeatureFlags contains feature flags configured in rill.yaml
	FeatureFlags map[string]string `db:"feature_flags"`
	// Annotations to enrich activity events (like usage tracking)
	Annotations map[string]string
	// Paths to expose over HTTP (defaults to ./public)
	PublicPaths []string `db:"public_paths"`
	// IgnoreInitialInvalidProjectError indicates whether to ignore an invalid project error when the instance is initially created.
	IgnoreInitialInvalidProjectError bool `db:"-"`
	// AIInstructions is extra context for LLM/AI features. Used to guide natural language question answering and routing.
	AIInstructions string `db:"ai_instructions"`
	// FrontendURL is the URL of the web interface.
	FrontendURL string `db:"frontend_url"`
}

// InstanceConfig contains dynamic configuration for an instance.
// It is configured by parsing instance variables prefixed with "rill.".
// For example, a variable "rill.stage_changes=true" would set the StageChanges field to true.
// InstanceConfig should only be used for config that the user is allowed to change dynamically at runtime.
type InstanceConfig struct {
	// DownloadLimitBytes is the limit on size of exported file. If set to 0, there is no limit.
	DownloadLimitBytes int64 `mapstructure:"rill.download_limit_bytes"`
	// InteractiveSQLRowLimit is the row limit for interactive SQL queries. It does not apply to exports of SQL queries. If set to 0, there is no limit.
	InteractiveSQLRowLimit int64 `mapstructure:"rill.interactive_sql_row_limit"`
	// StageChanges indicates whether to keep previously ingested tables for sources/models, and only override them if ingestion of a new table is successful.
	StageChanges bool `mapstructure:"rill.stage_changes"`
	// WatchRepo configures the project parser to setup a file watcher to instantly detect and parse changes to the project files.
	WatchRepo bool `mapstructure:"rill.watch_repo"`
	// ModelDefaultMaterialize indicates whether to materialize models by default.
	ModelDefaultMaterialize bool `mapstructure:"rill.models.default_materialize"`
	// ModelMaterializeDelaySeconds adds a delay before materializing models.
	ModelMaterializeDelaySeconds uint32 `mapstructure:"rill.models.materialize_delay_seconds"`
	// ModelConcurrentExecutionLimit sets the maximum number of concurrent model executions.
	ModelConcurrentExecutionLimit uint32 `mapstructure:"rill.models.concurrent_execution_limit"`
	// MetricsComparisonsExact indicates whether to rewrite metrics comparison queries to approximately correct queries.
	// Approximated comparison queries are faster but may not return comparison data points for all values.
	MetricsApproximateComparisons bool `mapstructure:"rill.metrics.approximate_comparisons"`
	// MetricsApproximateComparisonsCTE indicates whether to rewrite metrics comparison queries to use a CTE for base query.
	MetricsApproximateComparisonsCTE bool `mapstructure:"rill.metrics.approximate_comparisons_cte"`
	// MetricsApproxComparisonTwoPhaseLimit if query limit is less than this then rewrite metrics comparison queries to use a two-phase comparison approach where first query is used to get the base values and the second query is used to get the comparison values.
	MetricsApproxComparisonTwoPhaseLimit int64 `mapstructure:"rill.metrics.approximate_comparisons_two_phase_limit"`
	// MetricsExactifyDruidTopN indicates whether to split Druid TopN queries into two queries to increase the accuracy of the returned measures.
	// Enabling it reduces the performance of Druid toplist queries.
	// See runtime/metricsview/executor_rewrite_druid_exactify.go for more details.
	MetricsExactifyDruidTopN bool `mapstructure:"rill.metrics.exactify_druid_topn"`
	// MetricsNullFillingImplementation switches between null-filling implementations for timeseries queries.
	// Can be "", "none", "new", "pushdown".
	MetricsNullFillingImplementation string `mapstructure:"rill.metrics.timeseries_null_filling_implementation"`
	// AlertsDefaultStreamingRefreshCron sets a default cron expression for refreshing alerts with streaming refs.
	// Namely, this is used to check alerts against external tables (e.g. in Druid) where new data may be added at any time (i.e. is considered "streaming").
	AlertsDefaultStreamingRefreshCron string `mapstructure:"rill.alerts.default_streaming_refresh_cron"`
	// AlertsFastStreamingRefreshCron is similar to AlertsDefaultStreamingRefreshCron but is used for alerts that are based on always-on OLAP connectors (i.e. that have MayScaleToZero == false).
	AlertsFastStreamingRefreshCron string `mapstructure:"rill.alerts.fast_streaming_refresh_cron"`
}

// ResolveOLAPConnector resolves the OLAP connector to default to for the instance.
func (i *Instance) ResolveOLAPConnector() string {
	if i.ProjectOLAPConnector != "" {
		return i.ProjectOLAPConnector
	}
	if i.OLAPConnector != "" {
		return i.OLAPConnector
	}
	// Fallback to duckdb for backwards compatibility with projects that don't specify an OLAP connector
	return "duckdb"
}

func (i *Instance) ResolveAIConnector() string {
	if i.ProjectAIConnector != "" {
		return i.ProjectAIConnector
	}
	return i.AIConnector
}

// ResolveVariables returns the final resolved variables
func (i *Instance) ResolveVariables(withLowerKeys bool) map[string]string {
	r := make(map[string]string, len(i.ProjectVariables)+len(i.Variables))

	// set ProjectVariables first i.e. Project defaults
	for k, v := range i.ProjectVariables {
		if withLowerKeys {
			k = strings.ToLower(k)
		}
		r[k] = v
	}

	// override with instance Variables
	for k, v := range i.Variables {
		if withLowerKeys {
			k = strings.ToLower(k)
		}
		r[k] = v
	}
	return r
}

// Config resolves the current dynamic config properties for the instance.
// See InstanceConfig for details.
func (i *Instance) Config() (InstanceConfig, error) {
	// Default config
	res := InstanceConfig{
		DownloadLimitBytes:                   int64(datasize.MB * 128),
		InteractiveSQLRowLimit:               10_000,
		StageChanges:                         true,
		WatchRepo:                            i.Environment == "dev",
		ModelDefaultMaterialize:              false,
		ModelMaterializeDelaySeconds:         0,
		ModelConcurrentExecutionLimit:        5,
		MetricsApproximateComparisons:        true,
		MetricsApproximateComparisonsCTE:     false,
		MetricsApproxComparisonTwoPhaseLimit: 250,
		MetricsExactifyDruidTopN:             false,
		AlertsDefaultStreamingRefreshCron:    "0 0 * * *",    // Every 24 hours
		AlertsFastStreamingRefreshCron:       "*/10 * * * *", // Every 10 minutes
	}

	// Resolve variables
	vars := i.ResolveVariables(true)

	// Backwards compatibility: Use "__materialize_default" as alias for "rill.models.default_materialize".
	if vars != nil {
		if v, ok := vars["__materialize_default"]; ok {
			if _, ok := vars["rill.models.default_materialize"]; !ok {
				vars["rill.models.default_materialize"] = v
			}
		}
	}

	// Decode variables into res.
	err := mapstructure.WeakDecode(vars, &res)
	if err != nil {
		return InstanceConfig{}, fmt.Errorf("failed to parse instance config: %w", err)
	}

	return res, nil
}

func (i *Instance) ResolveConnectors() []*runtimev1.Connector {
	var res []*runtimev1.Connector
	res = append(res, i.Connectors...)
	res = append(res, i.ProjectConnectors...)
	// implicit connectors
	vars := i.ResolveVariables(true)
	for k := range vars {
		if !strings.HasPrefix(k, "connector.") {
			continue
		}

		parts := strings.Split(k, ".")
		if len(parts) <= 2 {
			continue
		}

		// Implicitly defined connectors always have the same name as the driver
		name := parts[1]
		found := false
		for _, c := range res {
			if c.Name == name {
				found = true
				break
			}
		}
		if !found {
			res = append(res, &runtimev1.Connector{
				Type: name,
				Name: name,
			})
		}
	}
	return res
}
