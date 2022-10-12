package drivers

import (
	"context"
	"time"
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
	CatalogObjectTypeSource         string = "source"
	CatalogObjectTypeMetricsView    string = "metrics_view"
	CatalogObjectTypeUnmanagedTable string = "unmanaged_table"
)

// CatalogObject represents one object in the catalog, such as a source
type CatalogObject struct {
	Name      string
	Type      string
	SQL       string
	CreatedOn time.Time `db:"created_on"`
	UpdatedOn time.Time `db:"updated_on"`
}
