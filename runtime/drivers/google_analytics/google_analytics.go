package google_analytics

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/gcputil"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"google.golang.org/api/analyticsdata/v1beta"
	"google.golang.org/api/option"
)

func init() {
	drivers.Register("google_analytics", driver{})
	drivers.RegisterAsConnector("google_analytics", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Google Analytics 4",
	Description: "Import data from Google Analytics 4.",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "google_application_credentials",
			Type:        drivers.FilePropertyType,
			DisplayName: "GCP Credentials",
			Description: "GCP credentials as JSON string",
			Placeholder: "Paste your GCP service account JSON here",
			Secret:      true,
			Required:    true,
		},
		{
			Key:         "property_id",
			Type:        drivers.StringPropertyType,
			DisplayName: "GA4 Property ID",
			Description: "Google Analytics 4 property ID",
			Placeholder: "123456789",
			Required:    true,
		},
	},
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "report_type",
			Type:        drivers.StringPropertyType,
			DisplayName: "Report type",
			Description: "Predefined report or custom dimensions/metrics",
			Required:    true,
		},
		{
			Key:         "dimensions",
			Type:        drivers.StringPropertyType,
			DisplayName: "Dimensions",
			Description: "Comma-separated GA4 dimension names (for custom reports)",
		},
		{
			Key:         "metrics",
			Type:        drivers.StringPropertyType,
			DisplayName: "Metrics",
			Description: "Comma-separated GA4 metric names (for custom reports)",
		},
		{
			Key:         "start_date",
			Type:        drivers.StringPropertyType,
			DisplayName: "Start date",
			Description: "Start date in YYYY-MM-DD format",
			Required:    true,
		},
		{
			Key:         "end_date",
			Type:        drivers.StringPropertyType,
			DisplayName: "End date",
			Description: "End date in YYYY-MM-DD format (defaults to today)",
		},
		{
			Key:         "name",
			Type:        drivers.StringPropertyType,
			DisplayName: "Source name",
			Description: "The name of the source",
			Placeholder: "my_ga4_source",
			Required:    true,
		},
	},
	ImplementsWarehouse: false,
}

type driver struct{}

type configProperties struct {
	SecretJSON      string `mapstructure:"google_application_credentials"`
	PropertyID      string `mapstructure:"property_id"`
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("google_analytics driver can't be shared")
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	return &connection{
		config: conf,
		logger: logger,
	}, nil
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

type connection struct {
	config *configProperties
	logger *zap.Logger
}

var _ drivers.Handle = &connection{}

// Ping validates the credentials and property access by running a minimal report.
func (c *connection) Ping(ctx context.Context) error {
	svc, err := c.createService(ctx)
	if err != nil {
		return fmt.Errorf("failed to create GA4 service: %w", err)
	}

	// Run a minimal report to validate credentials and property access
	req := &analyticsdata.RunReportRequest{
		DateRanges: []*analyticsdata.DateRange{{StartDate: "yesterday", EndDate: "today"}},
		Metrics:    []*analyticsdata.Metric{{Name: "activeUsers"}},
		Limit:      1,
	}

	propertyID := fmt.Sprintf("properties/%s", c.config.PropertyID)
	_, err = svc.Properties.RunReport(propertyID, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to access GA4 property %s: %w", c.config.PropertyID, err)
	}

	return nil
}

func (c *connection) createService(ctx context.Context) (*analyticsdata.Service, error) {
	creds, err := gcputil.Credentials(ctx, c.config.SecretJSON, c.config.AllowHostAccess, analyticsdata.AnalyticsReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}
	svc, err := analyticsdata.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create analyticsdata service: %w", err)
	}
	return svc, nil
}

// Driver implements drivers.Handle.
func (c *connection) Driver() string {
	return "google_analytics"
}

// Config implements drivers.Handle.
func (c *connection) Config() map[string]any {
	return maps.Clone(map[string]any{
		"google_application_credentials": c.config.SecretJSON,
		"property_id":                    c.config.PropertyID,
		"allow_host_access":              c.config.AllowHostAccess,
	})
}

// Close implements drivers.Handle.
func (c *connection) Close() error {
	return nil
}

// Migrate implements drivers.Handle.
func (c *connection) Migrate(ctx context.Context) error {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsRegistry implements drivers.Handle.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Handle.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Handle.
func (c *connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsAdmin implements drivers.Handle.
func (c *connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (c *connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// AsOLAP implements drivers.Handle.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsInformationSchema implements drivers.Handle.
func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Handle.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
}

// AsFileStore implements drivers.Handle.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return c, true
}

// AsNotifier implements drivers.Handle.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}
