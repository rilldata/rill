package server

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
		return nil, err
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
		return nil, err
	}

	err = s.admin.DB.UpdateSuperuser(ctx, user.ID, req.Superuser)
	if err != nil {
		return nil, err
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
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
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
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	if req.Preferences.TimeZone != nil {
		_, err := time.LoadLocation(*req.Preferences.TimeZone)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid time zone: %s", *req.Preferences.TimeZone)
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
		DisplayName:          user.DisplayName,
		PhotoURL:             user.PhotoURL,
		GithubUsername:       user.GithubUsername,
		GithubToken:          user.GithubToken,
		GithubTokenExpiresOn: user.GithubTokenExpiresOn,
		GithubRefreshToken:   user.GithubRefreshToken,
		QuotaSingleuserOrgs:  user.QuotaSingleuserOrgs,
		QuotaTrialOrgs:       user.QuotaTrialOrgs,
		PreferenceTimeZone:   valOrDefault(req.Preferences.TimeZone, user.PreferenceTimeZone),
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

func (s *Server) ListUserAuthTokens(ctx context.Context, req *adminv1.ListUserAuthTokensRequest) (*adminv1.ListUserAuthTokensResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.user_id", req.UserId),
	)

	claims := auth.GetClaims(ctx)
	forceAccess := claims.Superuser(ctx) && req.SuperuserForceAccess

	userID := req.UserId
	if userID == "current" { // Special alias for the current user
		if claims.OwnerType() != auth.OwnerTypeUser {
			return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
		}
		userID = claims.OwnerID()
	}
	if userID != claims.OwnerID() && !forceAccess {
		return nil, status.Error(codes.PermissionDenied, "not authorized to list auth tokens for other users")
	}

	pageSize := validPageSize(req.PageSize)
	pageToken, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	authTokens, err := s.admin.DB.FindUserAuthTokens(ctx, userID, pageToken.Val, pageSize, req.Refresh)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	if len(authTokens) >= pageSize {
		nextToken = marshalPageToken(authTokens[len(authTokens)-1].ID)
	}

	dtos := make([]*adminv1.UserAuthToken, len(authTokens))
	for i, t := range authTokens {
		var authClientID, authClientDisplayName, representingUserID string
		var expiresOn *timestamppb.Timestamp
		if t.AuthClientID != nil {
			authClientID = *t.AuthClientID
		}
		if t.AuthClientDisplayName != nil {
			authClientDisplayName = *t.AuthClientDisplayName
		}
		if t.RepresentingUserID != nil {
			representingUserID = *t.RepresentingUserID
		}
		if t.ExpiresOn != nil {
			expiresOn = timestamppb.New(*t.ExpiresOn)
		}

		id, err := uuid.Parse(t.ID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "invalid token ID %q: %v", t.ID, err)
		}

		prefix := authtoken.FromID(authtoken.TypeUser, id).Prefix()

		dtos[i] = &adminv1.UserAuthToken{
			Id:                    t.ID,
			DisplayName:           t.DisplayName,
			AuthClientId:          authClientID,
			AuthClientDisplayName: authClientDisplayName,
			RepresentingUserId:    representingUserID,
			CreatedOn:             timestamppb.New(t.CreatedOn),
			ExpiresOn:             expiresOn,
			UsedOn:                timestamppb.New(t.UsedOn),
			Prefix:                prefix,
			Refresh:               t.Refresh,
		}
	}

	return &adminv1.ListUserAuthTokensResponse{
		Tokens:        dtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) IssueUserAuthToken(ctx context.Context, req *adminv1.IssueUserAuthTokenRequest) (*adminv1.IssueUserAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.user_id", req.UserId),
		attribute.String("args.client_id", req.ClientId),
		attribute.String("args.display_name", req.DisplayName),
		attribute.Int64("args.ttl_minutes", req.TtlMinutes),
		attribute.Bool("args.has_represent_email", req.RepresentEmail != ""),
	)

	claims := auth.GetClaims(ctx)
	forceAccess := claims.Superuser(ctx) && req.SuperuserForceAccess

	userID := req.UserId
	if userID == "current" { // Special alias for the current user
		if claims.OwnerType() != auth.OwnerTypeUser {
			return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
		}
		userID = claims.OwnerID()
	}
	if userID != claims.OwnerID() && !forceAccess {
		return nil, status.Error(codes.PermissionDenied, "not authorized to issue auth tokens for other users")
	}

	var ttl *time.Duration
	if req.TtlMinutes > 0 {
		ttl = new(time.Duration)
		*ttl = time.Duration(req.TtlMinutes) * time.Minute
	}

	var representingUserID *string
	if req.RepresentEmail != "" {
		if !forceAccess {
			return nil, status.Error(codes.PermissionDenied, "not authorized to represent other users")
		}
		u, err := s.admin.DB.FindUserByEmail(ctx, req.RepresentEmail)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "user with email %q not found", req.RepresentEmail)
		}
		if u.ID == userID {
			return nil, status.Error(codes.InvalidArgument, "cannot represent yourself")
		}
		representingUserID = &u.ID
	}

	authToken, err := s.admin.IssueUserAuthToken(ctx, userID, req.ClientId, req.DisplayName, representingUserID, ttl, false)
	if err != nil {
		return nil, err
	}

	return &adminv1.IssueUserAuthTokenResponse{
		Token: authToken.Token().String(),
	}, nil
}

func (s *Server) RevokeUserAuthToken(ctx context.Context, req *adminv1.RevokeUserAuthTokenRequest) (*adminv1.RevokeUserAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.token_id", req.TokenId),
	)

	var tokenID string
	if req.TokenId == "current" { // Special alias for the current token
		if auth.GetClaims(ctx).OwnerType() != auth.OwnerTypeUser {
			return nil, status.Error(codes.PermissionDenied, "not authenticated with a user token")
		}
		tokenID = auth.GetClaims(ctx).AuthTokenID()
	} else {
		token, err := s.findUserAuthTokenFuzzy(ctx, req.TokenId)
		if err != nil {
			return nil, err
		}

		claims := auth.GetClaims(ctx)
		forceAccess := claims.Superuser(ctx) && req.SuperuserForceAccess
		if token.UserID != claims.OwnerID() && !forceAccess {
			return nil, status.Error(codes.PermissionDenied, "not authorized to revoke auth tokens for other users")
		}

		tokenID = token.ID
	}

	err := s.admin.DB.DeleteUserAuthToken(ctx, tokenID)
	if err != nil {
		return nil, err
	}

	return &adminv1.RevokeUserAuthTokenResponse{}, nil
}

func (s *Server) RevokeAllUserAuthTokens(ctx context.Context, req *adminv1.RevokeAllUserAuthTokensRequest) (*adminv1.RevokeAllUserAuthTokensResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.user_id", req.UserId),
	)

	claims := auth.GetClaims(ctx)
	forceAccess := claims.Superuser(ctx) && req.SuperuserForceAccess

	userID := req.UserId
	if userID == "current" { // Special alias for the current user
		if claims.OwnerType() != auth.OwnerTypeUser {
			return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
		}
		userID = claims.OwnerID()
	}
	if userID != claims.OwnerID() && !forceAccess {
		return nil, status.Error(codes.PermissionDenied, "not authorized to revoke auth tokens for other users")
	}

	tokensRevoked, err := s.admin.DB.DeleteAllUserAuthTokens(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &adminv1.RevokeAllUserAuthTokensResponse{
		TokensRevoked: int32(tokensRevoked),
	}, nil
}

func (s *Server) RevokeRepresentativeAuthTokens(ctx context.Context, req *adminv1.RevokeRepresentativeAuthTokensRequest) (*adminv1.RevokeRepresentativeAuthTokensResponse, error) {
	claims := auth.GetClaims(ctx)

	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can manage representative auth tokens")
	}

	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	u, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	observability.AddRequestAttributes(ctx,
		attribute.String("args.user_id", u.ID),
	)

	err = s.admin.DB.DeleteUserAuthTokensByUserAndRepresentingUser(ctx, claims.OwnerID(), u.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.RevokeRepresentativeAuthTokensResponse{}, nil
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
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	u, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	observability.AddRequestAttributes(ctx,
		attribute.String("args.user_id", u.ID),
	)

	ttl := time.Duration(req.TtlMinutes) * time.Minute
	displayName := fmt.Sprintf("Support for %s", u.Email)

	token, err := s.admin.IssueUserAuthToken(ctx, claims.OwnerID(), database.AuthClientIDRillSupport, displayName, &u.ID, &ttl, false)
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
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}
	tokenID := claims.AuthTokenID()

	err := s.admin.DB.DeleteUserAuthToken(ctx, tokenID)
	if err != nil {
		return nil, err
	}

	return &adminv1.RevokeCurrentAuthTokenResponse{}, nil
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
		res.Resource = &adminv1.SudoGetResourceResponse_Org{Org: s.organizationToDTO(org, true)}
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

func (s *Server) DeleteUser(ctx context.Context, req *adminv1.DeleteUserRequest) (*adminv1.DeleteUserResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.email", req.Email))

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find user by email: %v", err)
	}

	claims := auth.GetClaims(ctx)
	isCurrentUser := claims.OwnerType() == auth.OwnerTypeUser && claims.OwnerID() == user.ID
	forceAccess := claims.Superuser(ctx) && req.SuperuserForceAccess
	if !isCurrentUser && !forceAccess {
		return nil, status.Error(codes.PermissionDenied, "you can only delete your own user unless you are a superuser")
	}

	err = s.admin.DB.DeleteUser(ctx, user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &adminv1.DeleteUserResponse{}, nil
}

func (s *Server) SudoUpdateUserQuotas(ctx context.Context, req *adminv1.SudoUpdateUserQuotasRequest) (*adminv1.SudoUpdateUserQuotasResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.email", req.Email))
	if req.SingleuserOrgs != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.singleuser_orgs", int(*req.SingleuserOrgs)))
	}
	if req.TrialOrgs != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.trial_orgs", int(*req.TrialOrgs)))
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
		DisplayName:          user.DisplayName,
		PhotoURL:             user.PhotoURL,
		GithubUsername:       user.GithubUsername,
		GithubToken:          user.GithubToken,
		GithubTokenExpiresOn: user.GithubTokenExpiresOn,
		GithubRefreshToken:   user.GithubRefreshToken,
		QuotaSingleuserOrgs:  int(valOrDefault(req.SingleuserOrgs, int32(user.QuotaSingleuserOrgs))),
		QuotaTrialOrgs:       int(valOrDefault(req.TrialOrgs, int32(user.QuotaTrialOrgs))),
		PreferenceTimeZone:   user.PreferenceTimeZone,
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.SudoUpdateUserQuotasResponse{User: userToPB(updatedUser)}, nil
}

// SearchProjectUsers returns a list of users that match the given search/email query.
func (s *Server) SearchProjectUsers(ctx context.Context, req *adminv1.SearchProjectUsersRequest) (*adminv1.SearchProjectUsersResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.email_query", req.EmailQuery),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
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
			SingleuserOrgs: int32(u.QuotaSingleuserOrgs),
			TrialOrgs:      int32(u.QuotaTrialOrgs),
		},
		CreatedOn: timestamppb.New(u.CreatedOn),
		UpdatedOn: timestamppb.New(u.UpdatedOn),
	}
}

func orgMemberUserToPB(m *database.OrganizationMemberUser) *adminv1.OrganizationMemberUser {
	var attributes *structpb.Struct
	if len(m.Attributes) > 0 {
		if s, err := structpb.NewStruct(m.Attributes); err == nil {
			attributes = s
		}
	}

	return &adminv1.OrganizationMemberUser{
		UserId:          m.ID,
		UserEmail:       m.Email,
		UserName:        m.DisplayName,
		UserPhotoUrl:    m.PhotoURL,
		RoleName:        m.RoleName,
		ProjectsCount:   uint32(m.ProjectsCount),
		UsergroupsCount: uint32(m.UsergroupsCount),
		CreatedOn:       timestamppb.New(m.CreatedOn),
		UpdatedOn:       timestamppb.New(m.UpdatedOn),
		Attributes:      attributes,
	}
}

func projMemberUserToPB(m *database.ProjectMemberUser) *adminv1.ProjectMemberUser {
	return &adminv1.ProjectMemberUser{
		UserId:       m.ID,
		UserEmail:    m.Email,
		UserName:     m.DisplayName,
		UserPhotoUrl: m.PhotoURL,
		RoleName:     m.RoleName,
		OrgRoleName:  m.OrgRoleName,
		CreatedOn:    timestamppb.New(m.CreatedOn),
		UpdatedOn:    timestamppb.New(m.UpdatedOn),
	}
}

func usergroupMemberUserToPB(m *database.UsergroupMemberUser) *adminv1.UsergroupMemberUser {
	return &adminv1.UsergroupMemberUser{
		UserId:       m.ID,
		UserEmail:    m.Email,
		UserName:     m.DisplayName,
		UserPhotoUrl: m.PhotoURL,
		CreatedOn:    timestamppb.New(m.CreatedOn),
		UpdatedOn:    timestamppb.New(m.UpdatedOn),
	}
}

func orgInviteToPB(i *database.OrganizationInviteWithRole) *adminv1.OrganizationInvite {
	return &adminv1.OrganizationInvite{
		Email:     i.Email,
		RoleName:  i.RoleName,
		InvitedBy: i.InvitedBy,
	}
}

func projInviteToPB(i *database.ProjectInviteWithRole) *adminv1.ProjectInvite {
	return &adminv1.ProjectInvite{
		Email:       i.Email,
		RoleName:    i.RoleName,
		OrgRoleName: i.OrgRoleName,
		InvitedBy:   i.InvitedBy,
	}
}

// findUserAuthTokenFuzzy attempts to find a user auth token by exact ID, full token string, or unique prefix.
func (s *Server) findUserAuthTokenFuzzy(ctx context.Context, input string) (*database.UserAuthToken, error) {
	claims := auth.GetClaims(ctx)
	userID := claims.OwnerID()

	// Try exact ID match
	token, err := s.admin.DB.FindUserAuthToken(ctx, input)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	if err == nil {
		return token, nil
	}

	// Try full token string
	tokenStr, err := authtoken.FromString(input)
	if err == nil {
		token, err := s.admin.DB.FindUserAuthToken(ctx, tokenStr.ID.String())
		if err != nil && !errors.Is(err, database.ErrNotFound) {
			return nil, err
		}
		if err == nil {
			return token, nil
		}
	}

	// Validate input length and prefix
	if len(input) < 10 || !strings.HasPrefix(input, "rill_usr_") {
		return nil, status.Error(codes.InvalidArgument, "invalid token ID (must be at least 10 characters and start with 'rill_usr_')")
	}

	// Find all tokens for the user and match by prefix
	dbTokens, err := s.admin.DB.FindUserAuthTokens(ctx, userID, "", 1000, nil)
	if err != nil {
		return nil, err
	}

	tokens := make([]*authtoken.Token, len(dbTokens))
	for i, dbToken := range dbTokens {
		id, err := uuid.Parse(dbToken.ID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "invalid token ID %q: %v", dbToken.ID, err)
		}
		tokens[i] = authtoken.FromID(authtoken.TypeUser, id)
	}

	matches := authtoken.MatchByPrefix(input, tokens)

	if len(matches) > 1 {
		return nil, status.Error(codes.InvalidArgument, "multiple tokens match the given prefix (please use the full ID)")
	}
	if len(matches) == 1 {
		for _, dbToken := range dbTokens {
			id, err := uuid.Parse(dbToken.ID)
			if err != nil {
				continue
			}
			if id == matches[0].ID {
				return dbToken, nil
			}
		}
	}
	return nil, status.Error(codes.NotFound, "token not found")
}
