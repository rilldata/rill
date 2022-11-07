package drivers

import (
	"context"
	"time"

	"github.com/rilldata/rill/runtime/api"
)

// CatalogStore is implemented by drivers capable of storing catalog info for a specific instance
type CatalogStore interface {
	FindObjects(ctx context.Context, instanceID string) []*CatalogObject
	FindObject(ctx context.Context, instanceID string, name string) (*CatalogObject, bool)
	CreateObject(ctx context.Context, instanceID string, object *CatalogObject) error
	UpdateObject(ctx context.Context, instanceID string, object *CatalogObject) error
	DeleteObject(ctx context.Context, instanceID string, name string) error
}

// Constants representing different kinds of catalog objects
const (
	CatalogObjectTypeSource      string = "source"
	CatalogObjectTypeModel       string = "model"
	CatalogObjectTypeMetricsView string = "metrics_view"
	CatalogObjectTypeTable       string = "table"
)

// CatalogObject represents one object in the catalog, such as a source
type CatalogObject struct {
	Name        string
	Type        string
	SQL         string
	Definition  []byte
	Path        string
	RefreshedOn time.Time `db:"refreshed_on"`
	CreatedOn   time.Time `db:"created_on"`
	UpdatedOn   time.Time `db:"updated_on"`
}

type CatalogObjectDefinition struct {
	// source specific fields
	Connector  string
	Properties map[string]any

	// model specific fields
	Dialect string

	// metrics view specific fields
	From          string
	TimeDimension string
	TimeGrains    []string
	Dimensions    []*api.MetricsView_Dimension
	Measures      []*api.MetricsView_Measure
}
