package server_test

import (
	"context"
	"strings"
	"testing"
	"time"

	adminpkg "github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestMagicAuthTokenPassword(t *testing.T) {
	ctx := context.Background()
	fix := testadmin.New(t)

	// Create an admin user with an org and a project.
	_, admin := fix.NewUser(t)
	org, err := admin.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
	require.NoError(t, err)
	proj, err := admin.CreateProject(ctx, &adminv1.CreateProjectRequest{
		Org:        org.Organization.Name,
		Project:    "proj1",
		ProdSlots:  1,
		SkipDeploy: true,
	})
	require.NoError(t, err)
	getProjectReq := &adminv1.GetProjectRequest{Org: org.Organization.Name, Project: proj.Project.Name}

	// Passwords that are too short are rejected.
	_, err = admin.IssueMagicAuthToken(ctx, &adminv1.IssueMagicAuthTokenRequest{
		Org:       org.Organization.Name,
		Project:   proj.Project.Name,
		Resources: []*adminv1.ResourceName{{Type: "rill.runtime.v1.Explore", Name: "explore1"}},
		Password:  "short",
	})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	// Issue a password-protected token.
	issueRes, err := admin.IssueMagicAuthToken(ctx, &adminv1.IssueMagicAuthTokenRequest{
		Org:         org.Organization.Name,
		Project:     proj.Project.Name,
		Resources:   []*adminv1.ResourceName{{Type: "rill.runtime.v1.Explore", Name: "explore1"}},
		DisplayName: "Protected",
		Password:    "correct horse",
	})
	require.NoError(t, err)

	// The token itself is locked: it reveals nothing and grants no access.
	locked := fix.NewClient(t, issueRes.Token)
	cur, err := locked.GetCurrentMagicAuthToken(ctx, &adminv1.GetCurrentMagicAuthTokenRequest{})
	require.NoError(t, err)
	require.True(t, cur.Locked)
	require.True(t, cur.Token.PasswordProtected)
	require.Empty(t, cur.Token.DisplayName)
	require.Empty(t, cur.Token.Resources)
	require.Empty(t, cur.Token.Url)
	_, err = locked.GetProject(ctx, getProjectReq)
	require.Equal(t, codes.PermissionDenied, status.Code(err))

	// A wrong password doesn't unlock it.
	_, err = locked.UnlockMagicAuthToken(ctx, &adminv1.UnlockMagicAuthTokenRequest{Password: "wrong password"})
	require.Equal(t, codes.PermissionDenied, status.Code(err))

	// The right password returns a session that acts as the unlocked token.
	unlockRes, err := locked.UnlockMagicAuthToken(ctx, &adminv1.UnlockMagicAuthTokenRequest{Password: "correct horse"})
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(unlockRes.SessionToken, "rill_mgs_"))
	require.True(t, unlockRes.ExpiresOn.AsTime().After(time.Now()))
	session := fix.NewClient(t, unlockRes.SessionToken)
	cur, err = session.GetCurrentMagicAuthToken(ctx, &adminv1.GetCurrentMagicAuthTokenRequest{})
	require.NoError(t, err)
	require.False(t, cur.Locked)
	require.True(t, cur.Token.PasswordProtected)
	require.Equal(t, "Protected", cur.Token.DisplayName)
	require.Len(t, cur.Token.Resources, 1)
	require.Equal(t, issueRes.Url, cur.Token.Url)
	getRes, err := session.GetProject(ctx, getProjectReq)
	require.NoError(t, err)
	require.Equal(t, proj.Project.Id, getRes.Project.Id)

	// A session can't be used to unlock the token again.
	_, err = session.UnlockMagicAuthToken(ctx, &adminv1.UnlockMagicAuthTokenRequest{Password: "correct horse"})
	require.Equal(t, codes.PermissionDenied, status.Code(err))

	// The list reports that the token is password-protected.
	listRes, err := admin.ListMagicAuthTokens(ctx, &adminv1.ListMagicAuthTokensRequest{Org: org.Organization.Name, Project: proj.Project.Name})
	require.NoError(t, err)
	require.Len(t, listRes.Tokens, 1)
	require.True(t, listRes.Tokens[0].PasswordProtected)
	tokenID := listRes.Tokens[0].Id

	// Setting and removing the password at the same time is invalid.
	_, err = admin.UpdateMagicAuthToken(ctx, &adminv1.UpdateMagicAuthTokenRequest{TokenId: tokenID, Password: proto.String("another password"), RemovePassword: true})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	// A project member who didn't create the token can't update it.
	viewerUser, viewer := fix.NewUser(t)
	_, err = admin.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
		Org:     org.Organization.Name,
		Project: proj.Project.Name,
		Email:   viewerUser.Email,
		Role:    database.ProjectRoleNameViewer,
	})
	require.NoError(t, err)
	_, err = viewer.UpdateMagicAuthToken(ctx, &adminv1.UpdateMagicAuthTokenRequest{TokenId: tokenID, RemovePassword: true})
	require.Equal(t, codes.PermissionDenied, status.Code(err))

	// Changing the password invalidates the old password immediately.
	// The locked client's claims still come from the auth cache (which holds the old hash) at this point,
	// which checks that the password is verified against the database rather than the cached token model.
	updRes, err := admin.UpdateMagicAuthToken(ctx, &adminv1.UpdateMagicAuthTokenRequest{TokenId: tokenID, Password: proto.String("new password 1")})
	require.NoError(t, err)
	require.True(t, updRes.Token.PasswordProtected)
	require.Equal(t, "Protected", updRes.Token.DisplayName)
	_, err = locked.UnlockMagicAuthToken(ctx, &adminv1.UnlockMagicAuthTokenRequest{Password: "correct horse"})
	require.Equal(t, codes.PermissionDenied, status.Code(err))
	unlockRes, err = locked.UnlockMagicAuthToken(ctx, &adminv1.UnlockMagicAuthTokenRequest{Password: "new password 1"})
	require.NoError(t, err)

	// Changing the password also invalidates existing sessions (once the auth cache has expired).
	fix.Admin.PurgeAuthTokenCache()
	_, err = session.GetProject(ctx, getProjectReq)
	require.Equal(t, codes.Unauthenticated, status.Code(err))
	session = fix.NewClient(t, unlockRes.SessionToken)
	_, err = session.GetProject(ctx, getProjectReq)
	require.NoError(t, err)

	// Removing the password makes the token itself usable again and drops its sessions.
	updRes, err = admin.UpdateMagicAuthToken(ctx, &adminv1.UpdateMagicAuthTokenRequest{TokenId: tokenID, RemovePassword: true, DisplayName: proto.String("Open")})
	require.NoError(t, err)
	require.False(t, updRes.Token.PasswordProtected)
	require.Equal(t, "Open", updRes.Token.DisplayName)
	_, _, err = fix.Admin.UnlockMagicAuthToken(ctx, tokenID, "new password 1", time.Hour)
	require.ErrorIs(t, err, adminpkg.ErrMagicAuthTokenNotPasswordProtected)
	fix.Admin.PurgeAuthTokenCache()
	cur, err = locked.GetCurrentMagicAuthToken(ctx, &adminv1.GetCurrentMagicAuthTokenRequest{})
	require.NoError(t, err)
	require.False(t, cur.Locked)
	require.False(t, cur.Token.PasswordProtected)
	require.Equal(t, "Open", cur.Token.DisplayName)
	_, err = locked.GetProject(ctx, getProjectReq)
	require.NoError(t, err)
	_, err = session.GetProject(ctx, getProjectReq)
	require.Equal(t, codes.Unauthenticated, status.Code(err))

	// Tokens without a password are unaffected.
	openRes, err := admin.IssueMagicAuthToken(ctx, &adminv1.IssueMagicAuthTokenRequest{
		Org:       org.Organization.Name,
		Project:   proj.Project.Name,
		Resources: []*adminv1.ResourceName{{Type: "rill.runtime.v1.Explore", Name: "explore1"}},
	})
	require.NoError(t, err)
	open := fix.NewClient(t, openRes.Token)
	cur, err = open.GetCurrentMagicAuthToken(ctx, &adminv1.GetCurrentMagicAuthTokenRequest{})
	require.NoError(t, err)
	require.False(t, cur.Locked)
	require.False(t, cur.Token.PasswordProtected)
	_, err = open.UnlockMagicAuthToken(ctx, &adminv1.UnlockMagicAuthTokenRequest{Password: "anything"})
	require.Equal(t, codes.PermissionDenied, status.Code(err))
}
