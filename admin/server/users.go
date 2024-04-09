package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	// Load time zone data for time.ParseInLocation
	_ "time/tzdata"
)

func (s *Server) ListSuperusers(ctx context.Context, req *adminv1.ListSuperusersRequest) (*adminv1.ListSuperusersResponse, error) {
	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can list superusers")
	}

	users, err := s.admin.DB.FindSuperusers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	dtos := make([]*adminv1.User, len(users))
	for i, user := range users {
		dtos[i] = userToPB(user)
	}

	return &adminv1.ListSuperusersResponse{Users: dtos}, nil
}

func (s *Server) SetSuperuser(ctx context.Context, req *adminv1.SetSuperuserRequest) (*adminv1.SetSuperuserResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.Bool("args.superuser", req.Superuser),
	)

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can add/remove superuser")
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, fmt.Errorf("user not found for email id %s", req.Email)
		}
		return nil, err
	}

	err = s.admin.DB.UpdateSuperuser(ctx, user.ID, req.Superuser)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.SetSuperuserResponse{}, nil
}

func (s *Server) SearchUsers(ctx context.Context, req *adminv1.SearchUsersRequest) (*adminv1.SearchUsersResponse, error) {
	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can search users by email")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	users, err := s.admin.DB.FindUsersByEmailPattern(ctx, req.EmailPattern, token.Val, pageSize)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	if len(users) >= pageSize {
		nextToken = marshalPageToken(users[len(users)-1].Email)
	}

	dtos := make([]*adminv1.User, len(users))
	for i, user := range users {
		dtos[i] = userToPB(user)
	}

	return &adminv1.SearchUsersResponse{
		Users:         dtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) GetCurrentUser(ctx context.Context, req *adminv1.GetCurrentUserRequest) (*adminv1.GetCurrentUserResponse, error) {
	// Return an empty result if not authenticated.
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() == auth.OwnerTypeAnon {
		return &adminv1.GetCurrentUserResponse{}, nil
	}

	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, fmt.Errorf("not authenticated as a user")
	}

	// Owner is a user
	u, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, err
	}

	return &adminv1.GetCurrentUserResponse{
		User: userToPB(u),
		Preferences: &adminv1.UserPreferences{
			TimeZone: &u.PreferenceTimeZone,
		},
	}, nil
}

func (s *Server) UpdateUserPreferences(ctx context.Context, req *adminv1.UpdateUserPreferencesRequest) (*adminv1.UpdateUserPreferencesResponse, error) {
	claims := auth.GetClaims(ctx)

	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, fmt.Errorf("not authenticated as a user")
	}

	if req.Preferences.TimeZone != nil {
		_, err := time.LoadLocation(*req.Preferences.TimeZone)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid time zone: %s", *req.Preferences.TimeZone))
		}

		observability.AddRequestAttributes(ctx, attribute.String("preferences_time_zone", *req.Preferences.TimeZone))
	}

	// Owner is a user
	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, err
	}

	// Update user quota here
	updatedUser, err := s.admin.DB.UpdateUser(ctx, user.ID, &database.UpdateUserOptions{
		DisplayName:         user.DisplayName,
		PhotoURL:            user.PhotoURL,
		GithubUsername:      user.GithubUsername,
		GithubRefreshToken:  user.GithubRefreshToken,
		QuotaSingleuserOrgs: user.QuotaSingleuserOrgs,
		PreferenceTimeZone:  valOrDefault(req.Preferences.TimeZone, user.PreferenceTimeZone),
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.UpdateUserPreferencesResponse{
		Preferences: &adminv1.UserPreferences{
			TimeZone: &updatedUser.PreferenceTimeZone,
		},
	}, nil
}

// IssueRepresentativeAuthToken returns the temporary auth token for representing email
func (s *Server) IssueRepresentativeAuthToken(ctx context.Context, req *adminv1.IssueRepresentativeAuthTokenRequest) (*adminv1.IssueRepresentativeAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.Int64("args.ttl_minutes", req.TtlMinutes),
	)

	claims := auth.GetClaims(ctx)

	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can search users by email")
	}

	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, fmt.Errorf("not authenticated as a user")
	}

	u, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	ttl := time.Duration(req.TtlMinutes) * time.Minute
	displayName := fmt.Sprintf("Support for %s", u.Email)

	token, err := s.admin.IssueUserAuthToken(ctx, claims.OwnerID(), database.AuthClientIDRillSupport, displayName, &u.ID, &ttl)
	if err != nil {
		return nil, err
	}

	return &adminv1.IssueRepresentativeAuthTokenResponse{
		Token: token.Token().String(),
	}, nil
}

// RevokeCurrentAuthToken revokes the current auth token
func (s *Server) RevokeCurrentAuthToken(ctx context.Context, req *adminv1.RevokeCurrentAuthTokenRequest) (*adminv1.RevokeCurrentAuthTokenResponse, error) {
	claims := auth.GetClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, fmt.Errorf("not authenticated as a user")
	}
	tokenID := claims.AuthTokenID()

	err := s.admin.DB.DeleteUserAuthToken(ctx, tokenID)
	if err != nil {
		return nil, err
	}

	return &adminv1.RevokeCurrentAuthTokenResponse{
		TokenId: tokenID,
	}, nil
}

func (s *Server) SudoGetResource(ctx context.Context, req *adminv1.SudoGetResourceRequest) (*adminv1.SudoGetResourceResponse, error) {
	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can lookup resource")
	}

	res := &adminv1.SudoGetResourceResponse{}
	switch id := req.Id.(type) {
	case *adminv1.SudoGetResourceRequest_UserId:
		user, err := s.admin.DB.FindUser(ctx, id.UserId)
		if err != nil {
			return nil, err
		}
		res.Resource = &adminv1.SudoGetResourceResponse_User{User: userToPB(user)}
	case *adminv1.SudoGetResourceRequest_OrgId:
		org, err := s.admin.DB.FindOrganization(ctx, id.OrgId)
		if err != nil {
			return nil, err
		}
		res.Resource = &adminv1.SudoGetResourceResponse_Org{Org: organizationToDTO(org)}
	case *adminv1.SudoGetResourceRequest_ProjectId:
		proj, err := s.admin.DB.FindProject(ctx, id.ProjectId)
		if err != nil {
			return nil, err
		}
		org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
		if err != nil {
			return nil, err
		}
		res.Resource = &adminv1.SudoGetResourceResponse_Project{Project: s.projToDTO(proj, org.Name)}
	case *adminv1.SudoGetResourceRequest_DeploymentId:
		depl, err := s.admin.DB.FindDeployment(ctx, id.DeploymentId)
		if err != nil {
			return nil, err
		}
		res.Resource = &adminv1.SudoGetResourceResponse_Deployment{Deployment: deploymentToDTO(depl)}
	case *adminv1.SudoGetResourceRequest_InstanceId:
		depl, err := s.admin.DB.FindDeploymentByInstanceID(ctx, id.InstanceId)
		if err != nil {
			return nil, err
		}
		res.Resource = &adminv1.SudoGetResourceResponse_Instance{Instance: deploymentToDTO(depl)}
	default:
		return nil, status.Errorf(codes.Internal, "unexpected resource type %T", id)
	}

	return res, nil
}

func (s *Server) GetUser(ctx context.Context, req *adminv1.GetUserRequest) (*adminv1.GetUserResponse, error) {
	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can get user")
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	return &adminv1.GetUserResponse{User: userToPB(user)}, nil
}

func (s *Server) SudoUpdateUserQuotas(ctx context.Context, req *adminv1.SudoUpdateUserQuotasRequest) (*adminv1.SudoUpdateUserQuotasResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.email", req.Email))
	if req.SingleuserOrgs != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.singleuser_orgs", int(*req.SingleuserOrgs)))
	}

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can manage quotas")
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	// Update user quota here
	updatedUser, err := s.admin.DB.UpdateUser(ctx, user.ID, &database.UpdateUserOptions{
		DisplayName:         user.DisplayName,
		PhotoURL:            user.PhotoURL,
		GithubUsername:      user.GithubUsername,
		GithubRefreshToken:  user.GithubRefreshToken,
		QuotaSingleuserOrgs: int(valOrDefault(req.SingleuserOrgs, uint32(user.QuotaSingleuserOrgs))),
		PreferenceTimeZone:  user.PreferenceTimeZone,
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.SudoUpdateUserQuotasResponse{User: userToPB(updatedUser)}, nil
}

// SearchProjectUsers returns a list of users that match the given search/email query.
func (s *Server) SearchProjectUsers(ctx context.Context, req *adminv1.SearchProjectUsersRequest) (*adminv1.SearchProjectUsersResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.email_query", req.EmailQuery),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject {
		return nil, status.Error(codes.PermissionDenied, "not authorized to search project users")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pageSize := validPageSize(req.PageSize)

	users, err := s.admin.DB.SearchProjectUsers(ctx, proj.ID, req.EmailQuery, token.Val, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	nextToken := ""
	if len(users) >= pageSize {
		nextToken = marshalPageToken(users[len(users)-1].Email)
	}

	dtos := make([]*adminv1.User, len(users))
	for i, user := range users {
		dtos[i] = userToPB(user)
	}

	return &adminv1.SearchProjectUsersResponse{
		Users:         dtos,
		NextPageToken: nextToken,
	}, nil
}

func userToPB(u *database.User) *adminv1.User {
	return &adminv1.User{
		Id:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		PhotoUrl:    u.PhotoURL,
		Quotas: &adminv1.UserQuotas{
			SingleuserOrgs: uint32(u.QuotaSingleuserOrgs),
		},
		CreatedOn: timestamppb.New(u.CreatedOn),
		UpdatedOn: timestamppb.New(u.UpdatedOn),
	}
}

func memberToPB(m *database.Member) *adminv1.Member {
	return &adminv1.Member{
		UserId:    m.ID,
		UserEmail: m.Email,
		UserName:  m.DisplayName,
		RoleName:  m.RoleName,
		CreatedOn: timestamppb.New(m.CreatedOn),
		UpdatedOn: timestamppb.New(m.UpdatedOn),
	}
}

func inviteToPB(i *database.Invite) *adminv1.UserInvite {
	return &adminv1.UserInvite{
		Email:     i.Email,
		Role:      i.Role,
		InvitedBy: i.InvitedBy,
	}
}
