package provisioner

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rilldata/rill/admin/database"
	"go.uber.org/zap"
)

type Provisioner interface {
	Provision(ctx context.Context, opts *ProvisionOptions) (*Allocation, error)
	Deprovision(ctx context.Context, provisionID string) error
	AwaitReady(ctx context.Context, provisionID string) error
	Update(ctx context.Context, provisionID string, newVersion string) error
	CheckCapacity(ctx context.Context) error
}

type ProvisionOptions struct {
	ProvisionID    string
	RuntimeVersion string
	Slots          int
	Annotations    map[string]string
}

type Allocation struct {
	Host         string
	Audience     string
	CPU          int
	MemoryGB     int
	StorageBytes int64
}

type ProvisionerSpec struct {
	Type string          `json:"type"`
	Spec json.RawMessage `json:"spec"`
}

func NewSet(set string, db database.DB, logger *zap.Logger) (map[string]Provisioner, error) {
	// Parse provisioner set
	pts := map[string]ProvisionerSpec{}
	err := json.Unmarshal([]byte(set), &pts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provisioner set: %w", err)
	}

	// Instantiate provisioners based on their type
	ps := make(map[string]Provisioner)
	for k, v := range pts {
		switch v.Type {
		case "static":
			p, err := NewStatic(v.Spec, db, logger)
			if err != nil {
				return nil, err
			}
			ps[k] = p
			continue
		case "kubernetes":
			p, err := NewKubernetes(v.Spec)
			if err != nil {
				return nil, err
			}
			ps[k] = p
			continue
		default:
			return nil, fmt.Errorf("invalid provisioner type %q", v.Type)
		}
	}

	return ps, nil
}
