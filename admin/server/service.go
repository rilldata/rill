package server

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/authtoken"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateService creates a new service account.
func (s *Server) CreateService(ctx context.Context, req *adminv1.CreateServiceRequest) (*adminv1.CreateServiceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.name", req.Name),
		attribute.String("args.organization", req.Org),
		attribute.String("args.org_role", req.OrgRoleName),
		attribute.String("args.project", req.Project),
		attribute.String("args.project_role", req.ProjectRoleName),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to create a service")
	}

	if req.OrgRoleName == "" && req.ProjectRoleName == "" {
		return nil, status.Error(codes.InvalidArgument, "at least one of org role or project role must be specified")
	}

	// Check if project name and role are both provided or both empty
	if (req.Project != "" && req.ProjectRoleName == "") || (req.Project == "" && req.ProjectRoleName != "") {
		return nil, status.Error(codes.InvalidArgument, "both project name and project role must be specified together")
	}

	ctx, tx, err := s.admin.DB.NewTx(ctx, true)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	opts := &database.InsertServiceOptions{
		OrgID:      org.ID,
		Name:       req.Name,
		Attributes: nil,
	}
	if req.Attributes != nil {
		opts.Attributes = req.Attributes.AsMap()
	}
	// Create service with attributes
	service, err := s.admin.DB.InsertService(ctx, opts)
	if err != nil {
		return nil, err
	}

	// If org role is specified, assign it
	if req.OrgRoleName != "" {
		orgRole, err := s.admin.DB.FindOrganizationRole(ctx, req.OrgRoleName)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid organization role")
		}
		err = s.admin.DB.InsertOrganizationMemberService(ctx, service.ID, org.ID, orgRole.ID)
		if err != nil {
			return nil, err
		}
	}

	// If project role is specified, assign it
	if req.Project != "" && req.ProjectRoleName != "" {
		project, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid project")
		}

		projectRole, err := s.admin.DB.FindProjectRole(ctx, req.ProjectRoleName)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid project role")
		}

		err = s.admin.DB.UpsertProjectMemberServiceRole(ctx, service.ID, project.ID, projectRole.ID)
		if err != nil {
			return nil, err
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, status.Error(codes.Internal, "failed to commit transaction")
	}

	return &adminv1.CreateServiceResponse{
		Service: serviceToPB(service, org.Name),
	}, nil
}

func (s *Server) GetService(ctx context.Context, req *adminv1.GetServiceRequest) (*adminv1.GetServiceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.name", req.Name),
		attribute.String("args.organization", req.Org),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to show service")
	}

	service, err := s.admin.DB.FindServiceByName(ctx, org.ID, req.Name)
	if err != nil {
		return nil, err
	}

	orgMemberService, err := s.admin.DB.FindOrganizationMemberServiceForService(ctx, service.ID)
	if err != nil {
		return nil, err
	}

	projectMemberServices, err := s.admin.DB.FindProjectMemberServicesForService(ctx, service.ID)
	if err != nil {
		return nil, err
	}
	var projectMemberServicesPB []*adminv1.ProjectMemberService
	for _, projectMemberService := range projectMemberServices {
		projectMemberServicesPB = append(projectMemberServicesPB, projectMemberServiceWithProjectToPB(projectMemberService, org.ID, org.Name))
	}

	return &adminv1.GetServiceResponse{
		Service:            orgMemberServiceToPB(orgMemberService, org.ID, org.Name),
		ProjectMemberships: projectMemberServicesPB,
	}, nil
}

// ListServices lists all service accounts in an organization.
func (s *Server) ListServices(ctx context.Context, req *adminv1.ListServicesRequest) (*adminv1.ListServicesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to list services")
	}

	services, err := s.admin.DB.FindOrganizationMemberServices(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var servicesPB []*adminv1.OrganizationMemberService
	for _, service := range services {
		servicesPB = append(servicesPB, orgMemberServiceToPB(service, org.ID, org.Name))
	}

	return &adminv1.ListServicesResponse{
		Services: servicesPB,
	}, nil
}

// ListProjectMemberServices lists all service accounts for a project.
func (s *Server) ListProjectMemberServices(ctx context.Context, req *adminv1.ListProjectMemberServicesRequest) (*adminv1.ListProjectMemberServicesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.project", req.Project),
		attribute.String("args.organization", req.Org),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}

	project, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to list project services")
	}

	services, err := s.admin.DB.FindProjectMemberServices(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	var servicesPB []*adminv1.ProjectMemberService
	for _, service := range services {
		servicesPB = append(servicesPB, projectMemberServiceToPB(service, org.ID, org.Name, project.ID, project.Name))
	}

	return &adminv1.ListProjectMemberServicesResponse{
		Services: servicesPB,
	}, nil
}

// UpdateService updates a service account.
func (s *Server) UpdateService(ctx context.Context, req *adminv1.UpdateServiceRequest) (*adminv1.UpdateServiceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.name", req.Name),
		attribute.String("args.organization", req.Org),
	)

	if req.NewName != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.new_name", *req.NewName))
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to update service")
	}

	service, err := s.admin.DB.FindServiceByName(ctx, org.ID, req.Name)
	if err != nil {
		return nil, err
	}

	updateOpts := &database.UpdateServiceOptions{
		Name:       valOrDefault(req.NewName, service.Name),
		Attributes: service.Attributes,
	}
	if req.Attributes != nil {
		updateOpts.Attributes = req.Attributes.AsMap()
	}

	service, err = s.admin.DB.UpdateService(ctx, service.ID, updateOpts)
	if err != nil {
		return nil, err
	}

	return &adminv1.UpdateServiceResponse{
		Service: serviceToPB(service, org.Name),
	}, nil
}

func (s *Server) SetOrganizationMemberServiceRole(ctx context.Context, req *adminv1.SetOrganizationMemberServiceRoleRequest) (*adminv1.SetOrganizationMemberServiceRoleResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.name", req.Name),
		attribute.String("args.organization", req.Org),
		attribute.String("args.role", req.Role),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	}

	service, err := s.admin.DB.FindServiceByName(ctx, org.ID, req.Name)
	if err != nil {
		return nil, err
	}

	orgRole, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.UpdateOrganizationMemberServiceRole(ctx, service.ID, org.ID, orgRole.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.SetOrganizationMemberServiceRoleResponse{}, nil
}

func (s *Server) RemoveOrganizationMemberService(ctx context.Context, req *adminv1.RemoveOrganizationMemberServiceRequest) (*adminv1.RemoveOrganizationMemberServiceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.name", req.Name),
		attribute.String("args.organization", req.Org),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove service")
	}

	service, err := s.admin.DB.FindServiceByName(ctx, org.ID, req.Name)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.DeleteOrganizationMemberService(ctx, service.ID, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RemoveOrganizationMemberServiceResponse{}, nil
}

func (s *Server) SetProjectMemberServiceRole(ctx context.Context, req *adminv1.SetProjectMemberServiceRoleRequest) (*adminv1.SetProjectMemberServiceRoleResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.name", req.Name),
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.role", req.Role),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	}

	project, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	service, err := s.admin.DB.FindServiceByName(ctx, org.ID, req.Name)
	if err != nil {
		return nil, err
	}

	projectRole, err := s.admin.DB.FindProjectRole(ctx, req.Role)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.UpsertProjectMemberServiceRole(ctx, service.ID, project.ID, projectRole.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.SetProjectMemberServiceRoleResponse{}, nil
}

func (s *Server) RemoveProjectMemberService(ctx context.Context, req *adminv1.RemoveProjectMemberServiceRequest) (*adminv1.RemoveProjectMemberServiceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.name", req.Name),
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}
	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove service from project")
	}

	project, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	service, err := s.admin.DB.FindServiceByName(ctx, org.ID, req.Name)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.DeleteProjectMemberService(ctx, service.ID, project.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.RemoveProjectMemberServiceResponse{}, nil
}

// DeleteService deletes a service account.
func (s *Server) DeleteService(ctx context.Context, req *adminv1.DeleteServiceRequest) (*adminv1.DeleteServiceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.name", req.Name),
		attribute.String("args.organization_name", req.Org),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}

	service, err := s.admin.DB.FindServiceByName(ctx, org.ID, req.Name)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to delete service")
	}

	err = s.admin.DB.DeleteService(ctx, service.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.DeleteServiceResponse{}, nil
}

// ListServiceAuthTokens lists all auth tokens for a service account.
func (s *Server) ListServiceAuthTokens(ctx context.Context, req *adminv1.ListServiceAuthTokensRequest) (*adminv1.ListServiceAuthTokensResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.service_name", req.ServiceName),
		attribute.String("args.organization_name", req.Org),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	}

	service, err := s.admin.DB.FindServiceByName(ctx, org.ID, req.ServiceName)
	if err != nil {
		return nil, err
	}

	tokens, err := s.admin.DB.FindServiceAuthTokens(ctx, service.ID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*adminv1.ServiceToken, len(tokens))
	for i, token := range tokens {
		id, err := uuid.Parse(token.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, "invalid token ID format")
		}

		prefix := authtoken.FromID(authtoken.TypeService, id).Prefix()

		dtos[i] = &adminv1.ServiceToken{
			Id:        token.ID,
			Prefix:    prefix,
			CreatedOn: timestamppb.New(token.CreatedOn),
			ExpiresOn: timestamppb.New(safeTime(token.ExpiresOn)),
		}
	}

	return &adminv1.ListServiceAuthTokensResponse{
		Tokens: dtos,
	}, nil
}

// IssueServiceAuthToken issues a new auth token for a service account.
func (s *Server) IssueServiceAuthToken(ctx context.Context, req *adminv1.IssueServiceAuthTokenRequest) (*adminv1.IssueServiceAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.service_name", req.ServiceName),
		attribute.String("args.organization_name", req.Org),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}

	service, err := s.admin.DB.FindServiceByName(ctx, org.ID, req.ServiceName)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	}

	token, err := s.admin.IssueServiceAuthToken(ctx, service.ID, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.IssueServiceAuthTokenResponse{
		Token: token.Token().String(),
	}, nil
}

// RevokServiceAuthToken revokes an auth token for a service account.
func (s *Server) RevokeServiceAuthToken(ctx context.Context, req *adminv1.RevokeServiceAuthTokenRequest) (*adminv1.RevokeServiceAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.token_id", req.TokenId),
	)

	token, err := s.admin.DB.FindServiceAuthToken(ctx, req.TokenId)
	if err != nil {
		return nil, err
	}

	service, err := s.admin.DB.FindService(ctx, token.ServiceID)
	if err != nil {
		return nil, err
	}

	org, err := s.admin.DB.FindOrganization(ctx, service.OrgID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to revoke auth token")
	}

	err = s.admin.DB.DeleteServiceAuthToken(ctx, token.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RevokeServiceAuthTokenResponse{}, nil
}

func serviceToPB(service *database.Service, orgName string) *adminv1.Service {
	attr, err := structpb.NewStruct(service.Attributes)
	if err != nil {
		panic(err)
	}
	return &adminv1.Service{
		Id:         service.ID,
		Name:       service.Name,
		OrgId:      service.OrgID,
		OrgName:    orgName,
		Attributes: attr,
		CreatedOn:  timestamppb.New(service.CreatedOn),
		UpdatedOn:  timestamppb.New(service.UpdatedOn),
	}
}

func orgMemberServiceToPB(service *database.OrganizationMemberService, orgID, orgName string) *adminv1.OrganizationMemberService {
	attr, err := structpb.NewStruct(service.Attributes)
	if err != nil {
		panic(err)
	}
	return &adminv1.OrganizationMemberService{
		Id:              service.ID,
		Name:            service.Name,
		OrgId:           orgID,
		OrgName:         orgName,
		RoleName:        service.RoleName,
		HasProjectRoles: service.HasProjectRoles,
		Attributes:      attr,
		CreatedOn:       timestamppb.New(service.CreatedOn),
		UpdatedOn:       timestamppb.New(service.UpdatedOn),
	}
}

func projectMemberServiceToPB(service *database.ProjectMemberService, orgID, orgName, projectID, projectName string) *adminv1.ProjectMemberService {
	attr, err := structpb.NewStruct(service.Attributes)
	if err != nil {
		panic(err)
	}
	return &adminv1.ProjectMemberService{
		Id:              service.ID,
		Name:            service.Name,
		OrgId:           orgID,
		OrgName:         orgName,
		OrgRoleName:     service.OrgRoleName,
		ProjectId:       projectID,
		ProjectName:     projectName,
		ProjectRoleName: service.RoleName,
		Attributes:      attr,
		CreatedOn:       timestamppb.New(service.CreatedOn),
		UpdatedOn:       timestamppb.New(service.UpdatedOn),
	}
}

func projectMemberServiceWithProjectToPB(service *database.ProjectMemberServiceWithProject, orgID, orgName string) *adminv1.ProjectMemberService {
	attr, err := structpb.NewStruct(service.Attributes)
	if err != nil {
		panic(err)
	}
	return &adminv1.ProjectMemberService{
		Id:              service.ID,
		Name:            service.Name,
		OrgId:           orgID,
		OrgName:         orgName,
		OrgRoleName:     service.OrgRoleName,
		ProjectId:       service.ProjectID,
		ProjectName:     service.ProjectName,
		ProjectRoleName: service.RoleName,
		Attributes:      attr,
		CreatedOn:       timestamppb.New(service.CreatedOn),
		UpdatedOn:       timestamppb.New(service.UpdatedOn),
	}
}

func safeTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
