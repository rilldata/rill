package drivers

import (
	"context"
	"errors"
	"time"
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

// ErrInconsistentControllerVersion is returned from Controller when an unexpected controller version is observed in the DB.
// An unexpected controller version will only be observed if multiple controllers are running simultanesouly (split brain).
var ErrInconsistentControllerVersion = errors.New("controller: inconsistent version")

// ErrResourceNotFound is returned from catalog functions when a referenced resource does not exist.
var ErrResourceNotFound = errors.New("controller: resource not found")

// ErrResourceAlreadyExists is returned from catalog functions when attempting to create a resource that already exists.
var ErrResourceAlreadyExists = errors.New("controller: resource already exists")

// CatalogStore is implemented by drivers capable of storing catalog info for a specific instance.
// Implementations should treat resource kinds as case sensitive, but resource names as case insensitive.
type CatalogStore interface {
	NextControllerVersion(ctx context.Context) (int64, error)
	CheckControllerVersion(ctx context.Context, v int64) error

	FindResources(ctx context.Context) ([]Resource, error)
	CreateResource(ctx context.Context, v int64, r Resource) error
	UpdateResource(ctx context.Context, v int64, r Resource) error
	DeleteResource(ctx context.Context, v int64, k, n string) error
	DeleteResources(ctx context.Context) error

	FindModelSplitsByKeys(ctx context.Context, modelID string, keys []string) ([]ModelSplit, error)
	FindModelSplitsByPending(ctx context.Context, modelID string, limit int) ([]ModelSplit, error)
	InsertModelSplit(ctx context.Context, modelID string, split ModelSplit) error
	UpdateModelSplit(ctx context.Context, modelID string, split ModelSplit) error
	DeleteModelSplits(ctx context.Context, modelID string) error
}

// Resource is an entry in a catalog store
type Resource struct {
	Kind string
	Name string
	Data []byte
}

// ModelSplit is a single executable unit of a model.
type ModelSplit struct {
	Key           string
	DataJSON      []byte
	DataUpdatedOn time.Time
	ExecutedOn    time.Time
	Error         string
	Elapsed       time.Duration
}
