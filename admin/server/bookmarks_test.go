package server_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	"github.com/rilldata/rill/cli/testcli"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
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
		res, err := admin.ListBookmarks(ctx, &adminv1.ListBookmarksRequest{ProjectId: projectID, ResourceKind: proto.String(exploreKind), ResourceName: proto.String("explore1")})
		require.NoError(t, err)
		require.Equal(t, []string{"Shared explore bookmark"}, bookmarkNames(res.Bookmarks))

		// The resource name is matched case-insensitively.
		res, err = viewer.ListBookmarks(ctx, &adminv1.ListBookmarksRequest{ProjectId: projectID, ResourceKind: proto.String(exploreKind), ResourceName: proto.String("EXPLORE1")})
		require.NoError(t, err)
		require.Equal(t, []string{"Shared explore bookmark", "Viewer personal explore bookmark"}, bookmarkNames(res.Bookmarks))

		// Filtering by kind only.
		res, err = admin.ListBookmarks(ctx, &adminv1.ListBookmarksRequest{ProjectId: projectID, ResourceKind: proto.String(canvasKind)})
		require.NoError(t, err)
		require.Equal(t, []string{"Admin personal canvas bookmark"}, bookmarkNames(res.Bookmarks))

		// A resource name without a kind is rejected.
		_, err = admin.ListBookmarks(ctx, &adminv1.ListBookmarksRequest{ProjectId: projectID, ResourceName: proto.String("explore1")})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("listing requires read access to the project", func(t *testing.T) {
		_, err := outsider.ListBookmarks(ctx, &adminv1.ListBookmarksRequest{ProjectId: projectID})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("users must pass a project ID", func(t *testing.T) {
		_, err := admin.ListBookmarks(ctx, &adminv1.ListBookmarksRequest{})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		_, err = admin.CreateBookmark(ctx, &adminv1.CreateBookmarkRequest{DisplayName: "No project", ResourceKind: exploreKind, ResourceName: "explore1"})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
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

	t.Run("members can't get each other's personal bookmarks", func(t *testing.T) {
		_, err := admin.GetBookmark(ctx, &adminv1.GetBookmarkRequest{BookmarkId: viewerPersonal.Bookmark.Id})
		require.Equal(t, codes.PermissionDenied, status.Code(err))

		// Shared bookmarks are visible to all members.
		res, err := viewer.GetBookmark(ctx, &adminv1.GetBookmarkRequest{BookmarkId: shared.Bookmark.Id})
		require.NoError(t, err)
		require.Equal(t, shared.Bookmark.Id, res.Bookmark.Id)
	})

	t.Run("members without manage permission can't share their bookmarks", func(t *testing.T) {
		_, err := viewer.UpdateBookmark(ctx, &adminv1.UpdateBookmarkRequest{
			BookmarkId:  viewerPersonal.Bookmark.Id,
			DisplayName: "Viewer personal explore bookmark",
			Shared:      true,
		})
		require.Equal(t, codes.PermissionDenied, status.Code(err))

		res, err := viewer.GetBookmark(ctx, &adminv1.GetBookmarkRequest{BookmarkId: viewerPersonal.Bookmark.Id})
		require.NoError(t, err)
		require.False(t, res.Bookmark.Shared)
	})

	t.Run("can_manage reflects ownership and the manage permission", func(t *testing.T) {
		def, err := admin.CreateBookmark(ctx, &adminv1.CreateBookmarkRequest{
			DisplayName:  "Default canvas bookmark",
			ProjectId:    projectID,
			ResourceKind: canvasKind,
			ResourceName: "canvas1",
			Default:      true,
			Shared:       true,
		})
		require.NoError(t, err)
		require.True(t, def.Bookmark.CanManage)

		canManage := func(t *testing.T, c *client.Client) map[string]bool {
			res, err := c.ListBookmarks(ctx, &adminv1.ListBookmarksRequest{ProjectId: projectID})
			require.NoError(t, err)
			m := make(map[string]bool)
			for _, b := range res.Bookmarks {
				m[b.DisplayName] = b.CanManage
			}
			return m
		}

		require.Equal(t, map[string]bool{
			"Admin personal canvas bookmark": true,
			"Shared explore bookmark":        true,
			"Default canvas bookmark":        true,
		}, canManage(t, admin))
		require.Equal(t, map[string]bool{
			"Viewer personal explore bookmark (renamed)": true,
			"Shared explore bookmark":                    false,
			"Default canvas bookmark":                    false,
		}, canManage(t, viewer))

		res, err := viewer.GetBookmark(ctx, &adminv1.GetBookmarkRequest{BookmarkId: def.Bookmark.Id})
		require.NoError(t, err)
		require.False(t, res.Bookmark.CanManage)
	})
}

// TestEmbedBookmarks tests that embed users with an external user ID can manage their own personal bookmarks,
// and that they are isolated from other embed users and from project members.
func TestEmbedBookmarks(t *testing.T) {
	testmode.Expensive(t)

	adm := testadmin.NewWithOptionalRuntime(t, true)

	// Create test users. Note: the first user becomes a superuser, so we let the second one own the org/project to avoid confounding effects.
	_, _ = adm.NewUser(t)
	_, u1Client := adm.NewUser(t)

	// Deploy an empty test project
	projectDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "rill.yaml"), []byte("olap_connector: duckdb\n"), 0644))
	owner := testcli.New(t, adm, u1Client.Token)
	res := owner.Run(t, "org", "create", "bookmark-test")
	require.Equal(t, 0, res.ExitCode, res.Output)
	res = owner.Run(t, "project", "deploy", "--interactive=false", "--org=bookmark-test", "--project=bookmark-project", "--path="+projectDir)
	require.Equal(t, 0, res.ExitCode, res.Output)
	proj, err := adm.Admin.DB.FindProjectByName(t.Context(), "bookmark-test", "bookmark-project")
	require.NoError(t, err)

	// embedClient calls GetIFrame and returns a client using the iframe's admin token.
	// If resource is set, the embed is limited to that explore.
	embedClient := func(t *testing.T, externalUserID, resource string) *client.Client {
		req := &adminv1.GetIFrameRequest{
			Org:            "bookmark-test",
			Project:        "bookmark-project",
			ExternalUserId: externalUserID,
		}
		if resource != "" {
			req.Resource = resource
		} else {
			req.Navigation = true
		}
		resp, err := u1Client.GetIFrame(t.Context(), req)
		require.NoError(t, err)
		u, err := url.Parse(resp.IframeSrc)
		require.NoError(t, err)
		tkn := u.Query().Get("admin_token")
		require.NotEmpty(t, tkn)
		return adm.NewClient(t, tkn)
	}

	create := func(t *testing.T, c *client.Client, name, explore string) (*adminv1.Bookmark, error) {
		resp, err := c.CreateBookmark(t.Context(), &adminv1.CreateBookmarkRequest{
			DisplayName:  name,
			ProjectId:    proj.ID,
			ResourceKind: runtime.ResourceKindExplore,
			ResourceName: explore,
			UrlSearch:    "?tr=P7D",
		})
		if err != nil {
			return nil, err
		}
		return resp.Bookmark, nil
	}

	get := func(t *testing.T, c *client.Client, id string) (*adminv1.Bookmark, error) {
		resp, err := c.GetBookmark(t.Context(), &adminv1.GetBookmarkRequest{BookmarkId: id})
		if err != nil {
			return nil, err
		}
		return resp.Bookmark, nil
	}

	update := func(t *testing.T, c *client.Client, id string, shared bool) error {
		_, err := c.UpdateBookmark(t.Context(), &adminv1.UpdateBookmarkRequest{BookmarkId: id, DisplayName: "Renamed", UrlSearch: "?tr=P1D", Shared: shared})
		return err
	}

	remove := func(t *testing.T, c *client.Client, id string) error {
		_, err := c.RemoveBookmark(t.Context(), &adminv1.RemoveBookmarkRequest{BookmarkId: id})
		return err
	}

	list := func(t *testing.T, c *client.Client) []string {
		resp, err := c.ListBookmarks(t.Context(), &adminv1.ListBookmarksRequest{ProjectId: proj.ID})
		require.NoError(t, err)
		ids := make([]string, len(resp.Bookmarks))
		for i, b := range resp.Bookmarks {
			ids[i] = b.Id
		}
		return ids
	}

	embedA := embedClient(t, "user-a", "")
	embedB := embedClient(t, "user-b", "")

	t.Run("embed user can manage their own bookmarks", func(t *testing.T) {
		b, err := create(t, embedA, "Mine", "e1")
		require.NoError(t, err)
		require.True(t, b.CanManage)
		require.Empty(t, b.UserId)
		require.Contains(t, list(t, embedA), b.Id)

		got, err := get(t, embedA, b.Id)
		require.NoError(t, err)
		require.True(t, got.CanManage)

		require.NoError(t, update(t, embedA, b.Id, false))
		got, err = get(t, embedA, b.Id)
		require.NoError(t, err)
		require.Equal(t, "Renamed", got.DisplayName)

		require.NoError(t, remove(t, embedA, b.Id))
		require.NotContains(t, list(t, embedA), b.Id)
	})

	t.Run("embed users can omit the project ID", func(t *testing.T) {
		resp, err := embedA.CreateBookmark(t.Context(), &adminv1.CreateBookmarkRequest{
			DisplayName:  "Without project ID",
			ResourceKind: runtime.ResourceKindExplore,
			ResourceName: "e1",
			UrlSearch:    "?tr=P7D",
		})
		require.NoError(t, err)
		require.Equal(t, proj.ID, resp.Bookmark.ProjectId)

		listResp, err := embedA.ListBookmarks(t.Context(), &adminv1.ListBookmarksRequest{})
		require.NoError(t, err)
		require.True(t, slices.ContainsFunc(listResp.Bookmarks, func(b *adminv1.Bookmark) bool { return b.Id == resp.Bookmark.Id }))

		require.NoError(t, remove(t, embedA, resp.Bookmark.Id))
	})

	t.Run("embed users can't access each other's bookmarks", func(t *testing.T) {
		b, err := create(t, embedA, "A's bookmark", "e1")
		require.NoError(t, err)
		require.NotContains(t, list(t, embedB), b.Id)
		_, err = get(t, embedB, b.Id)
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.Equal(t, codes.PermissionDenied, status.Code(update(t, embedB, b.Id, false)))
		require.Equal(t, codes.PermissionDenied, status.Code(remove(t, embedB, b.Id)))
	})

	t.Run("project members can't access embed users' bookmarks", func(t *testing.T) {
		b, err := create(t, embedA, "A's bookmark", "e1")
		require.NoError(t, err)
		require.NotContains(t, list(t, u1Client), b.Id)
		_, err = get(t, u1Client, b.Id)
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.Equal(t, codes.PermissionDenied, status.Code(update(t, u1Client, b.Id, false)))
		require.Equal(t, codes.PermissionDenied, status.Code(remove(t, u1Client, b.Id)))
	})

	t.Run("embed users can't access project members' bookmarks", func(t *testing.T) {
		personal, err := create(t, u1Client, "Member's bookmark", "e1")
		require.NoError(t, err)
		require.NotContains(t, list(t, embedA), personal.Id)
		_, err = get(t, embedA, personal.Id)
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.Equal(t, codes.PermissionDenied, status.Code(update(t, embedA, personal.Id, false)))
		require.Equal(t, codes.PermissionDenied, status.Code(remove(t, embedA, personal.Id)))
	})

	t.Run("embed users don't see shared or default bookmarks", func(t *testing.T) {
		shared, err := u1Client.CreateBookmark(t.Context(), &adminv1.CreateBookmarkRequest{DisplayName: "Shared", ProjectId: proj.ID, ResourceKind: runtime.ResourceKindExplore, ResourceName: "e1", Shared: true})
		require.NoError(t, err)
		def, err := u1Client.CreateBookmark(t.Context(), &adminv1.CreateBookmarkRequest{DisplayName: "Default", ProjectId: proj.ID, ResourceKind: runtime.ResourceKindExplore, ResourceName: "e1", Default: true, Shared: true})
		require.NoError(t, err)

		for _, id := range []string{shared.Bookmark.Id, def.Bookmark.Id} {
			require.NotContains(t, list(t, embedA), id)
			_, err = get(t, embedA, id)
			require.Equal(t, codes.PermissionDenied, status.Code(err))
			require.Equal(t, codes.PermissionDenied, status.Code(update(t, embedA, id, true)))
			require.Equal(t, codes.PermissionDenied, status.Code(remove(t, embedA, id)))
		}
	})

	t.Run("embed users can't create or share shared and default bookmarks", func(t *testing.T) {
		_, err := embedA.CreateBookmark(t.Context(), &adminv1.CreateBookmarkRequest{DisplayName: "Shared", ProjectId: proj.ID, ResourceKind: runtime.ResourceKindExplore, ResourceName: "e1", Shared: true})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		_, err = embedA.CreateBookmark(t.Context(), &adminv1.CreateBookmarkRequest{DisplayName: "Default", ProjectId: proj.ID, ResourceKind: runtime.ResourceKindExplore, ResourceName: "e1", Default: true})
		require.Equal(t, codes.PermissionDenied, status.Code(err))

		b, err := create(t, embedA, "Mine", "e1")
		require.NoError(t, err)
		require.Equal(t, codes.PermissionDenied, status.Code(update(t, embedA, b.Id, true)))
	})

	t.Run("embeds with fixed resources are limited to those resources", func(t *testing.T) {
		// The same embed user, once with navigation and once limited to e1.
		open := embedClient(t, "fixed", "")
		fixed := embedClient(t, "fixed", "e1")

		other, err := create(t, open, "Other", "e2")
		require.NoError(t, err)

		allowed, err := create(t, fixed, "Allowed", "e1")
		require.NoError(t, err)
		_, err = create(t, fixed, "Denied", "e2")
		require.Equal(t, codes.PermissionDenied, status.Code(err))

		require.Equal(t, []string{allowed.Id}, list(t, fixed))
		require.ElementsMatch(t, []string{allowed.Id, other.Id}, list(t, open))
		_, err = get(t, fixed, other.Id)
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.Equal(t, codes.PermissionDenied, status.Code(update(t, fixed, other.Id, false)))
		require.Equal(t, codes.PermissionDenied, status.Code(remove(t, fixed, other.Id)))
		require.NoError(t, update(t, fixed, allowed.Id, false))
		require.NoError(t, remove(t, fixed, allowed.Id))
	})

	t.Run("embed users can only create a limited number of bookmarks", func(t *testing.T) {
		c := embedClient(t, "cap", "")
		for i := 0; i < 100; i++ {
			_, err := create(t, c, fmt.Sprintf("Bookmark %d", i), "e1")
			require.NoError(t, err)
		}
		_, err := create(t, c, "One too many", "e1")
		require.Equal(t, codes.ResourceExhausted, status.Code(err))
	})

	t.Run("concurrent creates can't exceed the bookmark limit", func(t *testing.T) {
		c := embedClient(t, "concurrent-cap", "")
		for i := 0; i < 95; i++ {
			_, err := create(t, c, fmt.Sprintf("Bookmark %d", i), "e1")
			require.NoError(t, err)
		}

		var wg sync.WaitGroup
		errs := make([]error, 20)
		for i := range errs {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, errs[i] = create(t, c, fmt.Sprintf("Concurrent %d", i), "e1")
			}()
		}
		wg.Wait()

		var created int
		for _, err := range errs {
			if err == nil {
				created++
				continue
			}
			require.Equal(t, codes.ResourceExhausted, status.Code(err))
		}
		require.Equal(t, 5, created)
		require.Len(t, list(t, c), 100)
	})
}

func bookmarkNames(bookmarks []*adminv1.Bookmark) []string {
	names := make([]string, len(bookmarks))
	for i, b := range bookmarks {
		names[i] = b.DisplayName
	}
	return names
}
