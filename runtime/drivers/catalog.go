package drivers

import (
	"context"
	"time"

	"github.com/rilldata/rill/runtime/api"
)

// CatalogStore is implemented by drivers capable of storing catalog info for a specific instance
type CatalogStore interface {
	FindObjects(ctx context.Context, instanceID string, typ CatalogObjectType) []*CatalogObject
	FindObject(ctx context.Context, instanceID string, name string) (*CatalogObject, bool)
	CreateObject(ctx context.Context, instanceID string, object *CatalogObject) error
	UpdateObject(ctx context.Context, instanceID string, object *CatalogObject) error
	DeleteObject(ctx context.Context, instanceID string, name string) error
}

// CatalogObject represents one object in the catalog, such as a source
type CatalogObject struct {
	Name        string
	Type        CatalogObjectType
	SQL         string
	Schema      *api.StructType
	Managed     bool
	CreatedOn   time.Time `db:"created_on"`
	UpdatedOn   time.Time `db:"updated_on"`
	RefreshedOn time.Time `db:"refreshed_on"`
}

// Constants representing different kinds of catalog objects
type CatalogObjectType string

const (
	CatalogObjectTypeUnspecified CatalogObjectType = ""
	CatalogObjectTypeTable       CatalogObjectType = "table"
	CatalogObjectTypeSource      CatalogObjectType = "source"
	CatalogObjectTypeMetricsView CatalogObjectType = "metrics_view"
)
