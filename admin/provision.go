package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/provisioner"
)

type ProvisionOptions struct {
	DeploymentID string
	Type         provisioner.ResourceType
	Name         string
	Provisioner  string
	Args         map[string]any
	Annotations  map[string]string
}

func (s *Service) Provision(ctx context.Context, opts *ProvisionOptions) (*database.ProvisionerResource, error) {
	// Attempt to find an existing provisioned resource
	pr, err := s.DB.FindProvisionerResourceByTypeAndName(ctx, opts.DeploymentID, string(opts.Type), opts.Name)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}

	// Find the provisioner to use
	var provisionerName string
	var p provisioner.Provisioner
	if pr != nil {
		if opts.Provisioner != "" && opts.Provisioner != pr.Provisioner {
			return nil, fmt.Errorf("provisioner: cannot change provisioner from %q to %q for deployment %q", provisionerName, opts.Provisioner, opts.DeploymentID)
		}

		var ok bool
		provisionerName = pr.Provisioner
		p, ok = s.ProvisionerSet[provisionerName]
		if !ok {
			return nil, fmt.Errorf("provisioner: previous provisioner %q is no longer in the provisioner set", provisionerName)
		}

		if !p.Supports(opts.Type) {
			return nil, fmt.Errorf("provisioner: previous provisioner %q no longer supports resource type %q", provisionerName, opts.Type)
		}
	} else if opts.Provisioner != "" {
		provisionerName = opts.Provisioner
		var ok bool
		p, ok = s.ProvisionerSet[provisionerName]
		if !ok {
			return nil, fmt.Errorf("provisioner: the requested provisioner %q is not in the provisioner set", provisionerName)
		}

		if !p.Supports(opts.Type) {
			return nil, fmt.Errorf("provisioner: the requested provisioner %q does not support resource type %q", provisionerName, opts.Type)
		}
	} else {
		for n, candidate := range s.ProvisionerSet {
			if candidate.Supports(opts.Type) {
				provisionerName = n
				p = candidate
				break
			}
		}
		if p == nil {
			return nil, fmt.Errorf("provisioner: no provisioner available that supports resource type %q", opts.Type)
		}
	}

	// Insert a pending provisioner resource if it doesn't exist
	if pr == nil {
		pr, err = s.DB.InsertProvisionerResource(ctx, &database.InsertProvisionerResourceOptions{
			ID:            uuid.New().String(),
			DeploymentID:  opts.DeploymentID,
			Type:          string(opts.Type),
			Name:          opts.Name,
			Status:        database.ProvisionerResourceStatusPending,
			StatusMessage: "Provisioning...",
			Provisioner:   provisionerName,
			Args:          opts.Args,
			State:         nil, // Will be populated after provisioning
			Config:        nil, // Will be populated after provisioning
		})
		if err != nil {
			if !errors.Is(err, database.ErrNotUnique) {
				return nil, err
			}

			// The resource must have been created concurrently by another process, so we try to find it again.
			pr, err = s.DB.FindProvisionerResourceByTypeAndName(ctx, opts.DeploymentID, string(opts.Type), opts.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to find expected provisioner resource: %w", err)
			}
		}
	}

	// Provision the resource
	r := &provisioner.Resource{
		ID:     pr.ID,
		Type:   opts.Type,
		State:  pr.State,  // Empty if inserting
		Config: pr.Config, // Empty if inserting
	}
	r, err = p.Provision(ctx, r, &provisioner.ResourceOptions{
		Args:        opts.Args,
		Annotations: opts.Annotations,
		RillVersion: s.resolveRillVersion(),
	})
	if err != nil {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		_, _ = s.DB.UpdateProvisionerResource(ctx, pr.ID, &database.UpdateProvisionerResourceOptions{
			Status:        database.ProvisionerResourceStatusError,
			StatusMessage: fmt.Sprintf("Failed provisioning: %v", err),
			Args:          pr.Args,
			State:         pr.State,
			Config:        pr.Config,
		})
		return nil, err
	}

	// Update the provisioner resource
	pr, err = s.DB.UpdateProvisionerResource(ctx, pr.ID, &database.UpdateProvisionerResourceOptions{
		Status:        database.ProvisionerResourceStatusOK,
		StatusMessage: "",
		Args:          opts.Args,
		State:         r.State,
		Config:        r.Config,
	})
	if err != nil {
		return nil, err
	}

	// Await the resource to be ready
	err = p.AwaitReady(ctx, r)
	if err != nil {
		return nil, err
	}

	return pr, nil
}
