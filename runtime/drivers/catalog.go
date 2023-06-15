package drivers

import (
	"context"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/proto"
)

// Constants representing the kinds of catalog objects.
type ObjectType int

const (
	ObjectTypeUnspecified ObjectType = 0
	ObjectTypeTable       ObjectType = 1
	ObjectTypeSource      ObjectType = 2
	ObjectTypeModel       ObjectType = 3
	ObjectTypeMetricsView ObjectType = 4
)

// CatalogStore is implemented by drivers capable of storing catalog info for a specific instance.
type CatalogStore interface {
	FindEntries(ctx context.Context, instanceID string, t ObjectType) ([]*CatalogEntry, error)
	FindEntry(ctx context.Context, instanceID string, name string) (*CatalogEntry, error)
	CreateEntry(ctx context.Context, instanceID string, entry *CatalogEntry) error
	UpdateEntry(ctx context.Context, instanceID string, entry *CatalogEntry) error
	DeleteEntry(ctx context.Context, instanceID string, name string) error
	DeleteEntries(ctx context.Context, instanceID string) error
}

// CatalogEntry represents one object in the catalog, such as a source.
type CatalogEntry struct {
	Name          string
	Type          ObjectType
	Object        proto.Message
	Path          string
	Embedded      bool
	BytesIngested int64
	Parents       []string
	Children      []string
	CreatedOn     time.Time
	UpdatedOn     time.Time
	RefreshedOn   time.Time
}

func (e *CatalogEntry) GetTable() *runtimev1.Table {
	obj, ok := e.Object.(*runtimev1.Table)
	if !ok {
		panic(fmt.Errorf("entry '%s' is not a table", e.Name))
	}
	return obj
}

func (e *CatalogEntry) GetSource() *runtimev1.Source {
	obj, ok := e.Object.(*runtimev1.Source)
	if !ok {
		panic(fmt.Errorf("entry '%s' is not a source", e.Name))
	}
	return obj
}

func (e *CatalogEntry) GetModel() *runtimev1.Model {
	obj, ok := e.Object.(*runtimev1.Model)
	if !ok {
		panic(fmt.Errorf("entry '%s' is not a model", e.Name))
	}
	return obj
}

func (e *CatalogEntry) GetMetricsView() *runtimev1.MetricsView {
	obj, ok := e.Object.(*runtimev1.MetricsView)
	if !ok {
		panic(fmt.Errorf("entry '%s' is not a metrics view", e.Name))
	}
	return obj
}
