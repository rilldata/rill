package server_test

import (
	"context"
	"testing"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestBookmarks(t *testing.T) {
	ctx := context.Background()
	fix := testadmin.New(t)

	const exploreKind = "rill.runtime.v1.Explore"
	const canvasKind = "rill.runtime.v1.Canvas"

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
	projectID := proj.Project.Id

	// Add a viewer to the project.
	viewerUser, viewer := fix.NewUser(t)
	_, err = admin.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
		Org:     org.Organization.Name,
		Project: proj.Project.Name,
		Email:   viewerUser.Email,
		Role:    database.ProjectRoleNameViewer,
	})
	require.NoError(t, err)

	// A user that is not a member of the project.
	_, outsider := fix.NewUser(t)

	// Bookmarks across two dashboards: a shared one and a personal one by the admin, and a personal one by the viewer.
	shared, err := admin.CreateBookmark(ctx, &adminv1.CreateBookmarkRequest{
		DisplayName:  "Shared explore bookmark",
		ProjectId:    projectID,
		ResourceKind: exploreKind,
		ResourceName: "explore1",
		Shared:       true,
		UrlSearch:    "?tr=P7D",
	})
	require.NoError(t, err)
	_, err = admin.CreateBookmark(ctx, &adminv1.CreateBookmarkRequest{
		DisplayName:  "Admin personal canvas bookmark",
		ProjectId:    projectID,
		ResourceKind: canvasKind,
		ResourceName: "canvas1",
		UrlSearch:    "?tr=P1D",
	})
	require.NoError(t, err)
	viewerPersonal, err := viewer.CreateBookmark(ctx, &adminv1.CreateBookmarkRequest{
		DisplayName:  "Viewer personal explore bookmark",
		ProjectId:    projectID,
		ResourceKind: exploreKind,
		ResourceName: "explore1",
		UrlSearch:    "?tr=P30D",
	})
	require.NoError(t, err)

	t.Run("project-wide listing returns own, shared and default bookmarks in a stable order", func(t *testing.T) {
		res, err := admin.ListBookmarks(ctx, &adminv1.ListBookmarksRequest{ProjectId: projectID})
		require.NoError(t, err)
		require.Equal(t, []string{"Admin personal canvas bookmark", "Shared explore bookmark"}, bookmarkNames(res.Bookmarks))

		res, err = viewer.ListBookmarks(ctx, &adminv1.ListBookmarksRequest{ProjectId: projectID})
		require.NoError(t, err)
		require.Equal(t, []string{"Shared explore bookmark", "Viewer personal explore bookmark"}, bookmarkNames(res.Bookmarks))
	})

	t.Run("listing can be filtered by resource", func(t *testing.T) {
		res, err := admin.ListBookmarks(ctx, &adminv1.ListBookmarksRequest{ProjectId: projectID, ResourceKind: exploreKind, ResourceName: "explore1"})
		require.NoError(t, err)
		require.Equal(t, []string{"Shared explore bookmark"}, bookmarkNames(res.Bookmarks))

		// The resource name is matched case-insensitively.
		res, err = viewer.ListBookmarks(ctx, &adminv1.ListBookmarksRequest{ProjectId: projectID, ResourceKind: exploreKind, ResourceName: "EXPLORE1"})
		require.NoError(t, err)
		require.Equal(t, []string{"Shared explore bookmark", "Viewer personal explore bookmark"}, bookmarkNames(res.Bookmarks))

		// Filtering by kind only.
		res, err = admin.ListBookmarks(ctx, &adminv1.ListBookmarksRequest{ProjectId: projectID, ResourceKind: canvasKind})
		require.NoError(t, err)
		require.Equal(t, []string{"Admin personal canvas bookmark"}, bookmarkNames(res.Bookmarks))

		// A resource name without a kind is rejected.
		_, err = admin.ListBookmarks(ctx, &adminv1.ListBookmarksRequest{ProjectId: projectID, ResourceName: "explore1"})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("listing requires read access to the project", func(t *testing.T) {
		_, err := outsider.ListBookmarks(ctx, &adminv1.ListBookmarksRequest{ProjectId: projectID})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("updating a bookmark bumps updated_on", func(t *testing.T) {
		before, err := viewer.GetBookmark(ctx, &adminv1.GetBookmarkRequest{BookmarkId: viewerPersonal.Bookmark.Id})
		require.NoError(t, err)

		_, err = viewer.UpdateBookmark(ctx, &adminv1.UpdateBookmarkRequest{
			BookmarkId:  viewerPersonal.Bookmark.Id,
			DisplayName: "Viewer personal explore bookmark (renamed)",
			UrlSearch:   "?tr=P30D",
		})
		require.NoError(t, err)

		after, err := viewer.GetBookmark(ctx, &adminv1.GetBookmarkRequest{BookmarkId: viewerPersonal.Bookmark.Id})
		require.NoError(t, err)
		require.Equal(t, "Viewer personal explore bookmark (renamed)", after.Bookmark.DisplayName)
		require.True(t, after.Bookmark.UpdatedOn.AsTime().After(before.Bookmark.UpdatedOn.AsTime()))

		// A viewer cannot update a shared bookmark.
		_, err = viewer.UpdateBookmark(ctx, &adminv1.UpdateBookmarkRequest{
			BookmarkId:  shared.Bookmark.Id,
			DisplayName: "Hijacked",
			Shared:      true,
			UrlSearch:   "?tr=P7D",
		})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})
}

func bookmarkNames(bookmarks []*adminv1.Bookmark) []string {
	names := make([]string, len(bookmarks))
	for i, b := range bookmarks {
		names[i] = b.DisplayName
	}
	return names
}
