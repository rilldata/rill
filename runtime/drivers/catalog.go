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

	FindModelSplits(ctx context.Context, opts *FindModelSplitsOptions) ([]ModelSplit, error)
	FindModelSplitsByKeys(ctx context.Context, modelID string, keys []string) ([]ModelSplit, error)
	CheckModelSplitsHaveErrors(ctx context.Context, modelID string) (bool, error)
	InsertModelSplit(ctx context.Context, modelID string, split ModelSplit) error
	UpdateModelSplit(ctx context.Context, modelID string, split ModelSplit) error
	UpdateModelSplitPending(ctx context.Context, modelID, splitKey string) error
	UpdateModelSplitsPendingIfError(ctx context.Context, modelID string) error
	DeleteModelSplits(ctx context.Context, modelID string) error

	FindInstanceHealth(ctx context.Context, instanceID string) (*InstanceHealth, error)
	UpsertInstanceHealth(ctx context.Context, h *InstanceHealth) error
}

// Resource is an entry in a catalog store
type Resource struct {
	Kind string
	Name string
	Data []byte
}

// ModelSplit represents a single executable unit of a model.
// Splits are an advanced feature that enables splitting and parallelizing execution of a model.
type ModelSplit struct {
	// Key is a unique identifier for the split. It should be a hash of DataJSON.
	Key string
	// DataJSON is the serialized parameters of the split.
	DataJSON []byte
	// Index is used to order the execution of splits.
	// Since it's just a guide and execution order usually is not critical,
	// it's okay if it's not unique or not always correct (e.g. for incrementally computed splits).
	Index int
	// Watermark represents the time when the underlying data that the split references was last updated.
	// If a split's watermark advances, we automatically schedule it for re-execution.
	Watermark *time.Time
	// ExecutedOn is the time when the split was last executed. If it is nil, the split is considered pending.
	ExecutedOn *time.Time
	// Error is the last error that occurred when executing the split.
	Error string
	// Elapsed is the duration of the last execution of the split.
	Elapsed time.Duration
}

// FindModelSplitsOptions is used to filter model splits.
type FindModelSplitsOptions struct {
	ModelID      string
	Limit        int
	WherePending bool
	WhereErrored bool
	AfterIndex   int
	AfterKey     string
}

// InstanceHealth represents the health of an instance.
type InstanceHealth struct {
	InstanceID string    `db:"instance_id"`
	HealthJSON []byte    `db:"health_json"`
	UpdatedOn  time.Time `db:"updated_on"`
}
