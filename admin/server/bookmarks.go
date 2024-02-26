package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Listbookmarks server returns the bookmarks for the user per project
func (s *Server) ListBookmarks(ctx context.Context, req *adminv1.ListBookmarksRequest) (*adminv1.ListBookmarksResponse, error) {
	claims := auth.GetClaims(ctx)
	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, fmt.Errorf("not authenticated as a user")
	}

	bookmarks, err := s.admin.DB.FindBookmarks(ctx, req.ProjectId, claims.OwnerID())
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
		return nil, fmt.Errorf("not authenticated as a user")
	}

	bookmark, err := s.admin.DB.FindBookmark(ctx, req.BookmarkId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, bookmark.ProjectID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if proj.Public {
		permissions.ReadProject = true
		permissions.ReadProd = true
	}

	if !permissions.ReadProject && !claims.Superuser(ctx) {
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
		return nil, fmt.Errorf("not authenticated as a user")
	}

	proj, err := s.admin.DB.FindProject(ctx, req.ProjectId)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if proj.Public {
		permissions.ReadProject = true
		permissions.ReadProd = true
	}

	if !permissions.ManageProject {
		req.IsGlobal = false // only users that can manage the project can
	}

	if !permissions.ReadProject && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read the project")
	}

	bookmark, err := s.admin.DB.InsertBookmark(ctx, &database.InsertBookmarkOptions{
		DisplayName:   req.DisplayName,
		Data:          req.Data,
		DashboardName: req.DashboardName,
		ProjectID:     req.ProjectId,
		UserID:        claims.OwnerID(),
		IsGlobal:      req.IsGlobal,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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
		return nil, fmt.Errorf("not authenticated as a user")
	}

	bookmark, err := s.admin.DB.FindBookmark(ctx, req.BookmarkId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, bookmark.ProjectID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)

	if !permissions.ManageProject {
		req.IsGlobal = false // only users that can manage the project can
	}

	if (!req.IsGlobal || !bookmark.IsGlobal) && bookmark.UserID != claims.OwnerID() {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to delete the bookmark")
	}

	err = s.admin.DB.UpdateBookmark(ctx, &database.UpdateBookmarkOptions{
		BookmarkID:  bookmark.ID,
		DisplayName: req.DisplayName,
		Data:        req.Data,
		IsGlobal:    req.IsGlobal,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.UpdateBookmarkResponse{}, nil
}

// RemoveBookmark server removes a bookmark for bookmark id
func (s *Server) RemoveBookmark(ctx context.Context, req *adminv1.RemoveBookmarkRequest) (*adminv1.RemoveBookmarkResponse, error) {
	claims := auth.GetClaims(ctx)

	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, fmt.Errorf("not authenticated as a user")
	}

	bookmark, err := s.admin.DB.FindBookmark(ctx, req.BookmarkId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, bookmark.ProjectID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)

	if (!bookmark.IsGlobal && bookmark.UserID != claims.OwnerID()) || !permissions.ManageProject {
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
		Id:            u.ID,
		DisplayName:   u.DisplayName,
		Data:          u.Data,
		DashboardName: u.DashboardName,
		ProjectId:     u.ProjectID,
		UserId:        u.UserID,
		CreatedOn:     timestamppb.New(u.CreatedOn),
		UpdatedOn:     timestamppb.New(u.UpdatedOn),
	}
}
