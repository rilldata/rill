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

	FindModelPartitions(ctx context.Context, opts *FindModelPartitionsOptions) ([]ModelPartition, error)
	FindModelPartitionsByKeys(ctx context.Context, modelID string, keys []string) ([]ModelPartition, error)
	CheckModelPartitionsHaveErrors(ctx context.Context, modelID string) (bool, error)
	InsertModelPartition(ctx context.Context, modelID string, partition ModelPartition) error
	UpdateModelPartition(ctx context.Context, modelID string, partition ModelPartition) error
	UpdateModelPartitionPending(ctx context.Context, modelID, partitionKey string) error
	UpdateModelPartitionsPendingIfError(ctx context.Context, modelID string) error
	DeleteModelPartitions(ctx context.Context, modelID string) error

	FindInstanceHealth(ctx context.Context, instanceID string) (*InstanceHealth, error)
	UpsertInstanceHealth(ctx context.Context, h *InstanceHealth) error
}

// Resource is an entry in a catalog store
type Resource struct {
	Kind string
	Name string
	Data []byte
}

// ModelPartition represents a single executable unit of a model.
// Partitions are an advanced feature that enables splitting and parallelizing execution of a model.
type ModelPartition struct {
	// Key is a unique identifier for the partition. It should be a hash of DataJSON.
	Key string
	// DataJSON is the serialized parameters of the partition.
	DataJSON []byte
	// Index is used to order the execution of partitions.
	// Since it's just a guide and execution order usually is not critical,
	// it's okay if it's not unique or not always correct (e.g. for incrementally computed partitions).
	Index int
	// Watermark represents the time when the underlying data that the partition references was last updated.
	// If a partition's watermark advances, we automatically schedule it for re-execution.
	Watermark *time.Time
	// ExecutedOn is the time when the partition was last executed. If it is nil, the partition is considered pending.
	ExecutedOn *time.Time
	// Error is the last error that occurred when executing the partition.
	Error string
	// Elapsed is the duration of the last execution of the partition.
	Elapsed time.Duration
}

// FindModelPartitionsOptions is used to filter model partitions.
type FindModelPartitionsOptions struct {
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
