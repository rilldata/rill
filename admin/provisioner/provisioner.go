package provisioner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rilldata/rill/admin/database"
	"go.uber.org/zap"
)

// ErrResourceTypeNotSupported should be returned by Provision if the provisioner does not support the requested resource type.
//
// By checking for this error, we can iterate over the chain of provisioners until we find a provisioner capable of provisioning the requested service.
var ErrResourceTypeNotSupported = errors.New("provisioner: resource type not supported")

// ProvisionerInitializer creates a new provisioner.
type ProvisionerInitializer func(specJSON []byte, db database.DB, logger *zap.Logger) (Provisioner, error)

// Initializers is a registry of provisioner initializers by type.
var Initializers = make(map[string]ProvisionerInitializer)

// Register registers a new provisioner initializer.
func Register(typ string, fn ProvisionerInitializer) {
	if Initializers[typ] != nil {
		panic(fmt.Errorf("already registered provisioner of type %q", typ))
	}
	Initializers[typ] = fn
}

// Provisioner is able to provision resources for one or more resource types.
type Provisioner interface {
	// Type returns the type of the provisioner.
	Type() string
	// Provision provisions a new resource.
	// It may be called multiple times for the same ID if:
	//  - the initial provision is interrupted, or
	//  - the resource args are updated
	//
	// This means Provision should be idempotent for the resource's ID (or otherwise do appropriate garbage collection in calls to Check).
	Provision(ctx context.Context, r *Resource, opts *ResourceOptions) (*Resource, error)
	// Deprovision deprovisions a resource.
	Deprovision(ctx context.Context, r *Resource) error
	// AwaitReady waits for a resource to be ready.
	AwaitReady(ctx context.Context, r *Resource) error
	// Check is called periodically to health check the provisioner.
	// The provided context should have a generous timeout to allow the provisioner to perform maintenance tasks.
	Check(ctx context.Context) error
	// CheckResource is called periodically to health check a specific resource.
	// The provided context should have a generous timeout to allow the provisioner to perform maintenance tasks for the resource.
	// The resource's state map will be updated to match that of the returned value.
	CheckResource(ctx context.Context, r *Resource, opts *ResourceOptions) (*Resource, error)
}

// ResourceOptions contains metadata about a resource.
type ResourceOptions struct {
	// Service-specific arguments for the provisioner. See resources.go for supported arguments.
	Args map[string]any
	// Annotations for the project the resource belongs to.
	Annotations map[string]string
	// RillVersion is the current version of Rill.
	RillVersion string
}

// Resource represents a provisioned resource.
type Resource struct {
	// ID uniquely identifies the provisioned resource.
	ID string
	// Type describes what type of service the resource is.
	Type ResourceType
	// Config contains access details for clients that use the resource.
	Config map[string]any
	// State contains state about the provisioned resource for use by the provisioner.
	// It should not be accessed outside of the provisioner.
	State map[string]any
}

// ProvisionerSpec is a JSON-serializable specification for a provisioner.
type ProvisionerSpec struct {
	Type string          `json:"type"`
	Spec json.RawMessage `json:"spec"`
}

// NewSet initializes a set of provisioners from a JSON specification.
// The JSON specification should be a map of names to ProvisionerSpecs.
func NewSet(setSpecJSON string, db database.DB, logger *zap.Logger) (map[string]Provisioner, error) {
	// Parse provisioner set
	pts := map[string]ProvisionerSpec{}
	err := json.Unmarshal([]byte(setSpecJSON), &pts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provisioner set: %w", err)
	}

	// Instantiate provisioners based on their type
	ps := make(map[string]Provisioner)
	for k, v := range pts {
		fn, ok := Initializers[v.Type]
		if !ok {
			return nil, fmt.Errorf("unknown type %q for provisioner %q", v.Type, k)
		}

		p, err := fn(v.Spec, db, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize provisioner %q: %w", k, err)
		}

		ps[k] = p
	}

	return ps, nil
}
