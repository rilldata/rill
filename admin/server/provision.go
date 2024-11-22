package server

import (
	"context"
	"errors"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/provisioner"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *Server) Provision(ctx context.Context, req *adminv1.ProvisionRequest) (*adminv1.ProvisionResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.deployment_id", req.DeploymentId),
		attribute.String("args.type", req.Type),
		attribute.String("args.name", req.Name),
	)

	// If the deployment ID is not provided, attempt to infer it from the access token.
	claims := auth.GetClaims(ctx)
	if req.DeploymentId == "" {
		if claims.OwnerType() == auth.OwnerTypeDeployment {
			req.DeploymentId = claims.OwnerID()
		} else {
			return nil, status.Error(codes.InvalidArgument, "missing deployment_id")
		}
	}

	depl, err := s.admin.DB.FindDeployment(ctx, req.DeploymentId)
	if err != nil {
		return nil, err
	}

	proj, err := s.admin.DB.FindProject(ctx, depl.ProjectID)
	if err != nil {
		return nil, err
	}

	permissions := auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ManageProvisionerResources {
		return nil, status.Error(codes.PermissionDenied, "not allowed to manage provisioner resources")
	}

	// If the resource is OK, return it immediately.
	res, err := s.admin.DB.FindProvisionerResourceByName(ctx, depl.ID, req.Type, req.Name)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	if res != nil && res.Status == database.ProvisionerResourceStatusOK {
		return &adminv1.ProvisionResponse{
			Resource: provisionerResourceToPB(res),
		}, nil
	}

	// Try or retry provisioning the resource.
	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}
	typ := provisioner.ResourceType(req.Type)
	if !typ.Valid() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid type %q", req.Type)
	}
	annotations := s.admin.NewDeploymentAnnotations(org, proj)
	res, err = s.admin.Provision(ctx, &admin.ProvisionOptions{
		DeploymentID: depl.ID,
		Type:         typ,
		Name:         req.Name,
		Provisioner:  "", // Means it should find a suitable provisioner
		Args:         req.Args.AsMap(),
		Annotations:  annotations.ToMap(),
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.ProvisionResponse{
		Resource: provisionerResourceToPB(res),
	}, nil
}

func provisionerResourceToPB(i *database.ProvisionerResource) *adminv1.ProvisionerResource {
	argsPB, err := structpb.NewStruct(i.Args)
	if err != nil {
		panic(err)
	}

	return &adminv1.ProvisionerResource{
		Id:           i.ID,
		DeploymentId: i.DeploymentID,
		Type:         string(i.Type),
		Name:         i.Name,
		Args:         argsPB,
	}
}
