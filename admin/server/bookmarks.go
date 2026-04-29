package server

import (
	"context"
	"errors"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListBookmarks server returns the bookmarks for the user per project
func (s *Server) ListBookmarks(ctx context.Context, req *adminv1.ListBookmarksRequest) (*adminv1.ListBookmarksResponse, error) {
	claims := auth.GetClaims(ctx)
	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	bookmarks, err := s.admin.DB.FindBookmarks(ctx, req.ProjectId, req.ResourceKind, req.ResourceName, claims.OwnerID())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	dtos := make([]*adminv1.Bookmark, len(bookmarks))
	for i, bookmark := range bookmarks {
		dtos[i] = bookmarkToPB(bookmark)
	}

	return &adminv1.ListBookmarksResponse{
		Bookmarks: dtos,
	}, nil
}

// GetBookmark server returns the bookmark for the user per project
func (s *Server) GetBookmark(ctx context.Context, req *adminv1.GetBookmarkRequest) (*adminv1.GetBookmarkResponse, error) {
	claims := auth.GetClaims(ctx)

	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	bookmark, err := s.admin.DB.FindBookmark(ctx, req.BookmarkId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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

	return &adminv1.GetBookmarkResponse{
		Bookmark: bookmarkToPB(bookmark),
	}, nil
}

// CreateBookmark server creates a bookmark for the user per project
func (s *Server) CreateBookmark(ctx context.Context, req *adminv1.CreateBookmarkRequest) (*adminv1.CreateBookmarkResponse, error) {
	claims := auth.GetClaims(ctx)

	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	proj, err := s.admin.DB.FindProject(ctx, req.ProjectId)
	if err != nil {
		return nil, err
	}

	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if proj.Public {
		permissions.CreateBookmarks = claims.OwnerType() == auth.OwnerTypeUser // Logged in users can create bookmarks on public projects
	}

	if !permissions.CreateBookmarks {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to create bookmarks")
	}

	if !permissions.ManageBookmarks && (req.Default || req.Shared) {
		// only admins can create shared/default bookmarks
		return nil, status.Error(codes.PermissionDenied, "does not have permission to create shared bookmarks")
	}

	if req.Default {
		// only one default bookmark can exist for a project/dashboard combo
		res, err := s.admin.DB.FindDefaultBookmark(ctx, req.ProjectId, req.ResourceKind, req.ResourceName)
		if err != nil && !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		}

		if res != nil {
			return nil, status.Error(codes.InvalidArgument, "default bookmark already exists")
		}
	}

	bookmark, err := s.admin.DB.InsertBookmark(ctx, &database.InsertBookmarkOptions{
		DisplayName:  req.DisplayName,
		Description:  req.Description,
		URLSearch:    req.UrlSearch,
		ResourceKind: req.ResourceKind,
		ResourceName: req.ResourceName,
		ProjectID:    req.ProjectId,
		UserID:       claims.OwnerID(),
		Default:      req.Default,
		Shared:       req.Shared,
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.CreateBookmarkResponse{
		Bookmark: bookmarkToPB(bookmark),
	}, nil
}

// UpdateBookmark updates a bookmark for the given user for the given project
func (s *Server) UpdateBookmark(ctx context.Context, req *adminv1.UpdateBookmarkRequest) (*adminv1.UpdateBookmarkResponse, error) {
	claims := auth.GetClaims(ctx)

	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	bookmark, err := s.admin.DB.FindBookmark(ctx, req.BookmarkId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, bookmark.ProjectID)
	if err != nil {
		return nil, err
	}

	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)

	if !permissions.ManageBookmarks && (bookmark.Shared || bookmark.Default) {
		// only admins can update shared/default bookmarks
		return nil, status.Error(codes.PermissionDenied, "does not have permission to update the bookmark")
	}

	if !bookmark.Shared && !bookmark.Default && bookmark.UserID != claims.OwnerID() {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to update the bookmark")
	}

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

// RemoveBookmark server removes a bookmark for bookmark id
func (s *Server) RemoveBookmark(ctx context.Context, req *adminv1.RemoveBookmarkRequest) (*adminv1.RemoveBookmarkResponse, error) {
	claims := auth.GetClaims(ctx)

	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	bookmark, err := s.admin.DB.FindBookmark(ctx, req.BookmarkId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, bookmark.ProjectID)
	if err != nil {
		return nil, err
	}

	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)

	if !permissions.ManageBookmarks && (bookmark.Shared || bookmark.Default) {
		// only admins can delete shared/default bookmarks
		return nil, status.Error(codes.PermissionDenied, "does not have permission to update the bookmark")
	}

	if !bookmark.Shared && !bookmark.Default && bookmark.UserID != claims.OwnerID() {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to delete the bookmark")
	}

	err = s.admin.DB.DeleteBookmark(ctx, req.BookmarkId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RemoveBookmarkResponse{}, nil
}

func bookmarkToPB(u *database.Bookmark) *adminv1.Bookmark {
	return &adminv1.Bookmark{
		Id:           u.ID,
		DisplayName:  u.DisplayName,
		Description:  u.Description,
		Data:         u.Data,
		UrlSearch:    u.URLSearch,
		ResourceKind: u.ResourceKind,
		ResourceName: u.ResourceName,
		ProjectId:    u.ProjectID,
		UserId:       u.UserID,
		Default:      u.Default,
		Shared:       u.Shared,
		CreatedOn:    timestamppb.New(u.CreatedOn),
		UpdatedOn:    timestamppb.New(u.UpdatedOn),
	}
}
