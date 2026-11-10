package server

import (
	"context"
	"errors"
	"strings"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// maxBookmarksPerEmbedUser caps the number of bookmarks an embed user can have in a project.
const maxBookmarksPerEmbedUser = 100

// ListBookmarks returns the bookmarks in a project that are visible to the user: their own plus shared and default ones.
// Embed users only see their own personal bookmarks.
// The resource kind and name are optional filters; when both are unset, all bookmarks in the project are returned.
func (s *Server) ListBookmarks(ctx context.Context, req *adminv1.ListBookmarksRequest) (*adminv1.ListBookmarksResponse, error) {
	claims := auth.GetClaims(ctx)
	o, err := ownerFromClaims(claims)
	if err != nil {
		return nil, err
	}

	if req.GetResourceName() != "" && req.GetResourceKind() == "" {
		return nil, status.Error(codes.InvalidArgument, "resource_kind is required when resource_name is set")
	}

	projectID, err := bookmarkProjectID(claims, req.ProjectId)
	if err != nil {
		return nil, err
	}

	proj, err := s.admin.DB.FindProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if proj.Public {
		permissions.ReadProject = true
		permissions.ReadProd = true
	}

	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read the project")
	}

	var bookmarks []*database.Bookmark
	var embedResources []database.ResourceName
	if o.isEmbed() {
		embedResources, err = embedResourcesFromClaims(claims)
		if err != nil {
			return nil, err
		}
		bookmarks, err = s.admin.DB.FindBookmarksByEmbedOwner(ctx, proj.ID, req.GetResourceKind(), req.GetResourceName(), o.embedUserID)
	} else {
		bookmarks, err = s.admin.DB.FindBookmarks(ctx, proj.ID, req.GetResourceKind(), req.GetResourceName(), o.userID)
	}
	if err != nil {
		return nil, err
	}

	dtos := make([]*adminv1.Bookmark, 0, len(bookmarks))
	for _, bookmark := range bookmarks {
		// Embeds limited to specific resources only list the bookmarks for those resources.
		if len(embedResources) > 0 && !containsResource(embedResources, bookmark.ResourceKind, bookmark.ResourceName) {
			continue
		}
		dtos = append(dtos, bookmarkToPB(bookmark, canManageBookmark(o, permissions, bookmark)))
	}

	return &adminv1.ListBookmarksResponse{
		Bookmarks: dtos,
	}, nil
}

// GetBookmark returns a bookmark that is visible to the caller: one of their own, or a shared or default one.
// Embed users can only get their own personal bookmarks.
func (s *Server) GetBookmark(ctx context.Context, req *adminv1.GetBookmarkRequest) (*adminv1.GetBookmarkResponse, error) {
	claims := auth.GetClaims(ctx)
	o, err := ownerFromClaims(claims)
	if err != nil {
		return nil, err
	}

	bookmark, err := s.admin.DB.FindBookmark(ctx, req.BookmarkId)
	if err != nil {
		return nil, err
	}

	proj, err := s.admin.DB.FindProject(ctx, bookmark.ProjectID)
	if err != nil {
		return nil, err
	}

	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if proj.Public {
		permissions.ReadProject = true
		permissions.ReadProd = true
	}

	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read the project")
	}

	owned := o.matches(bookmark.UserID, bookmark.EmbedOwnerID)
	if o.isEmbed() {
		// Shared and default bookmarks are not yet supported for embed users.
		if !owned {
			return nil, status.Error(codes.PermissionDenied, "does not have permission to read the bookmark")
		}
		if err := checkEmbedResource(claims, bookmark.ResourceKind, bookmark.ResourceName); err != nil {
			return nil, err
		}
	} else if !owned && !bookmark.Shared && !bookmark.Default {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read the bookmark")
	}

	return &adminv1.GetBookmarkResponse{
		Bookmark: bookmarkToPB(bookmark, canManageBookmark(o, permissions, bookmark)),
	}, nil
}

// CreateBookmark creates a bookmark owned by the caller.
// Only callers with the ManageBookmarks permission can create shared or default bookmarks.
func (s *Server) CreateBookmark(ctx context.Context, req *adminv1.CreateBookmarkRequest) (*adminv1.CreateBookmarkResponse, error) {
	claims := auth.GetClaims(ctx)
	o, err := ownerFromClaims(claims)
	if err != nil {
		return nil, err
	}

	projectID, err := bookmarkProjectID(claims, req.ProjectId)
	if err != nil {
		return nil, err
	}

	proj, err := s.admin.DB.FindProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if proj.Public && claims.OwnerType() == auth.OwnerTypeUser {
		// Logged in users can create bookmarks on public projects
		permissions.CreateBookmarks = true
	}

	if !permissions.CreateBookmarks {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to create bookmarks")
	}

	if !permissions.ManageBookmarks && (req.Default || req.Shared) {
		// only admins can create shared/default bookmarks
		return nil, status.Error(codes.PermissionDenied, "does not have permission to create shared bookmarks")
	}

	// Embed users are untrusted outside users, so we limit the resources they can bookmark and how many bookmarks they can create.
	if o.isEmbed() {
		if err := checkEmbedResource(claims, req.ResourceKind, req.ResourceName); err != nil {
			return nil, err
		}

		count, err := s.admin.DB.CountBookmarksByEmbedOwner(ctx, proj.ID, o.embedUserID)
		if err != nil {
			return nil, err
		}
		if count >= maxBookmarksPerEmbedUser {
			return nil, status.Errorf(codes.ResourceExhausted, "cannot create more than %d bookmarks", maxBookmarksPerEmbedUser)
		}
	}

	if req.Default {
		// only one default bookmark can exist for a project/dashboard combo
		res, err := s.admin.DB.FindDefaultBookmark(ctx, proj.ID, req.ResourceKind, req.ResourceName)
		if err != nil && !errors.Is(err, database.ErrNotFound) {
			return nil, err
		}

		if res != nil {
			return nil, status.Error(codes.AlreadyExists, "default bookmark already exists")
		}
	}

	bookmark, err := s.admin.DB.InsertBookmark(ctx, &database.InsertBookmarkOptions{
		DisplayName:  req.DisplayName,
		Description:  req.Description,
		URLSearch:    req.UrlSearch,
		ResourceKind: req.ResourceKind,
		ResourceName: req.ResourceName,
		ProjectID:    proj.ID,
		UserID:       o.userIDPtr(),
		EmbedOwnerID: o.embedUserIDPtr(),
		Default:      req.Default,
		Shared:       req.Shared,
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.CreateBookmarkResponse{
		Bookmark: bookmarkToPB(bookmark, canManageBookmark(o, permissions, bookmark)),
	}, nil
}

// UpdateBookmark updates a bookmark.
// Callers can update their own personal bookmarks; shared and default bookmarks require the ManageBookmarks permission.
func (s *Server) UpdateBookmark(ctx context.Context, req *adminv1.UpdateBookmarkRequest) (*adminv1.UpdateBookmarkResponse, error) {
	claims := auth.GetClaims(ctx)
	o, err := ownerFromClaims(claims)
	if err != nil {
		return nil, err
	}

	bookmark, err := s.admin.DB.FindBookmark(ctx, req.BookmarkId)
	if err != nil {
		return nil, err
	}

	proj, err := s.admin.DB.FindProject(ctx, bookmark.ProjectID)
	if err != nil {
		return nil, err
	}

	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if proj.Public {
		permissions.ReadProject = true
		permissions.ReadProd = true
	}

	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read the project")
	}

	if !canManageBookmark(o, permissions, bookmark) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to update the bookmark")
	}

	if o.isEmbed() {
		if err := checkEmbedResource(claims, bookmark.ResourceKind, bookmark.ResourceName); err != nil {
			return nil, err
		}
	}

	if req.Shared && !permissions.ManageBookmarks {
		// only admins can share bookmarks
		return nil, status.Error(codes.PermissionDenied, "does not have permission to share the bookmark")
	}

	// Note that req.Default is ignored: a bookmark's default flag can only be set on creation.
	err = s.admin.DB.UpdateBookmark(ctx, &database.UpdateBookmarkOptions{
		BookmarkID:  bookmark.ID,
		DisplayName: req.DisplayName,
		Description: req.Description,
		URLSearch:   req.UrlSearch,
		Shared:      req.Shared,
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.UpdateBookmarkResponse{}, nil
}

// RemoveBookmark removes a bookmark.
// Callers can remove their own personal bookmarks; shared and default bookmarks require the ManageBookmarks permission.
func (s *Server) RemoveBookmark(ctx context.Context, req *adminv1.RemoveBookmarkRequest) (*adminv1.RemoveBookmarkResponse, error) {
	claims := auth.GetClaims(ctx)
	o, err := ownerFromClaims(claims)
	if err != nil {
		return nil, err
	}

	bookmark, err := s.admin.DB.FindBookmark(ctx, req.BookmarkId)
	if err != nil {
		return nil, err
	}

	proj, err := s.admin.DB.FindProject(ctx, bookmark.ProjectID)
	if err != nil {
		return nil, err
	}

	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if proj.Public {
		permissions.ReadProject = true
		permissions.ReadProd = true
	}

	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read the project")
	}

	if !canManageBookmark(o, permissions, bookmark) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to delete the bookmark")
	}

	if o.isEmbed() {
		if err := checkEmbedResource(claims, bookmark.ResourceKind, bookmark.ResourceName); err != nil {
			return nil, err
		}
	}

	err = s.admin.DB.DeleteBookmark(ctx, req.BookmarkId)
	if err != nil {
		return nil, err
	}

	return &adminv1.RemoveBookmarkResponse{}, nil
}

// canManageBookmark returns true if the owner can update and remove the bookmark.
// Shared and default bookmarks require the ManageBookmarks permission; personal bookmarks can only be managed by their owner.
func canManageBookmark(o *owner, permissions *adminv1.ProjectPermissions, b *database.Bookmark) bool {
	if b.Shared || b.Default {
		return permissions.ManageBookmarks
	}
	return o.matches(b.UserID, b.EmbedOwnerID)
}

// bookmarkProjectID returns the project ID for a bookmark request.
// Embeds don't know their project's ID, so it defaults to the project of the embed token.
// An explicit project ID that doesn't match the embed token is rejected by the permission checks.
func bookmarkProjectID(claims auth.Claims, reqProjectID string) (string, error) {
	if reqProjectID != "" {
		return reqProjectID, nil
	}
	if tkn, ok := claims.AuthTokenModel().(*admin.EmbedToken); ok {
		return tkn.ProjectID, nil
	}
	return "", status.Error(codes.InvalidArgument, "project_id is required")
}

// embedResourcesFromClaims returns the resources an embed caller is limited to.
// An empty list means the embed is not limited to specific resources.
func embedResourcesFromClaims(claims auth.Claims) ([]database.ResourceName, error) {
	tkn, ok := claims.AuthTokenModel().(*admin.EmbedToken)
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected embed token model")
	}
	return tkn.Resources, nil
}

// checkEmbedResource returns a PermissionDenied error if the embed token is limited to specific resources and the given resource is not one of them.
func checkEmbedResource(claims auth.Claims, kind, name string) error {
	embedResources, err := embedResourcesFromClaims(claims)
	if err != nil {
		return err
	}
	if len(embedResources) > 0 && !containsResource(embedResources, kind, name) {
		return status.Errorf(codes.PermissionDenied, "this embed does not have access to bookmarks for %s %q", kind, name)
	}
	return nil
}

// containsResource returns true if the resource kind and name are in the list of resources.
// Resource names are case insensitive.
func containsResource(resources []database.ResourceName, kind, name string) bool {
	for _, r := range resources {
		if r.Type == kind && strings.EqualFold(r.Name, name) {
			return true
		}
	}
	return false
}

func bookmarkToPB(u *database.Bookmark, canManage bool) *adminv1.Bookmark {
	var userID string
	if u.UserID != nil {
		userID = *u.UserID
	}
	return &adminv1.Bookmark{
		Id:           u.ID,
		DisplayName:  u.DisplayName,
		Description:  u.Description,
		Data:         u.Data,
		UrlSearch:    u.URLSearch,
		ResourceKind: u.ResourceKind,
		ResourceName: u.ResourceName,
		ProjectId:    u.ProjectID,
		UserId:       userID,
		Default:      u.Default,
		Shared:       u.Shared,
		CanManage:    canManage,
		CreatedOn:    timestamppb.New(u.CreatedOn),
		UpdatedOn:    timestamppb.New(u.UpdatedOn),
	}
}
