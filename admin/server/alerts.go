package server

import (
	"context"
	"errors"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *Server) GetAlertMeta(ctx context.Context, req *adminv1.GetAlertMetaRequest) (*adminv1.GetAlertMetaResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.project_id", req.ProjectId),
		attribute.String("args.branch", req.Branch),
		attribute.String("args.alert", req.Alert),
		attribute.Bool("args.query_for", req.GetQueryFor() != nil),
	)

	proj, err := s.admin.DB.FindProject(ctx, req.ProjectId)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	permissions := auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProdStatus {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read alert meta")
	}

	if proj.ProdBranch != req.Branch {
		return nil, status.Error(codes.InvalidArgument, "branch not found")
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var attr map[string]any
	if req.QueryFor != nil {
		switch forVal := req.QueryFor.(type) {
		case *adminv1.GetAlertMetaRequest_QueryForUserId:
			attr, err = s.getAttributesForUser(ctx, proj.OrganizationID, proj.ID, forVal.QueryForUserId, "")
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		case *adminv1.GetAlertMetaRequest_QueryForUserEmail:
			attr, err = s.getAttributesForUser(ctx, proj.OrganizationID, proj.ID, "", forVal.QueryForUserEmail)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		default:
			return nil, status.Error(codes.InvalidArgument, "invalid 'for' type")
		}
	}

	var attrPB *structpb.Struct
	if attr != nil {
		attrPB, err = structpb.NewStruct(attr)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &adminv1.GetAlertMetaResponse{
		OpenUrl:            s.urls.alertOpen(org.Name, proj.Name, req.Alert),
		EditUrl:            s.urls.alertEdit(org.Name, proj.Name, req.Alert),
		QueryForAttributes: attrPB,
	}, nil
}
