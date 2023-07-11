package server

import (
	"context"
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
	// Return an empty result if not authenticated.
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() == auth.OwnerTypeAnon {
		return &adminv1.ListBookmarksResponse{}, nil
	}

	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, fmt.Errorf("not authenticated as a user")
	}

	bookmarks, err := s.admin.DB.FindBookmarks(ctx, req.ProjectId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	dtos := make([]*adminv1.DashboardBookmark, len(bookmarks))
	for i, bookmark := range bookmarks {
		dtos[i] = bookmarkToPB(bookmark)
	}

	return &adminv1.ListBookmarksResponse{
		DashboardBookmark: dtos,
	}, nil
}

// GetBookmark server returns the bookmark for the user per project
func (s *Server) GetBookmark(ctx context.Context, req *adminv1.GetBookmarkRequest) (*adminv1.GetBookmarkResponse, error) {
	// Return an empty result if not authenticated.
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() == auth.OwnerTypeAnon {
		return &adminv1.GetBookmarkResponse{}, nil
	}

	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, fmt.Errorf("not authenticated as a user")
	}

	bookmark, err := s.admin.DB.FindBookmark(ctx, req.BookmarkId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.GetBookmarkResponse{
		DashboardBookmark: bookmarkToPB(bookmark),
	}, nil
}

// CreateBookmark server creates a bookmark for the user per project
func (s *Server) CreateBookmark(ctx context.Context, req *adminv1.CreateBookmarkRequest) (*adminv1.CreateBookmarkResponse, error) {
	// Return an empty result if not authenticated.
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() == auth.OwnerTypeAnon {
		return &adminv1.CreateBookmarkResponse{}, nil
	}

	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, fmt.Errorf("not authenticated as a user")
	}

	bookmark, err := s.admin.DB.InsertBookmark(ctx, &database.InsertBookmarkOptions{
		DisplayName:   req.DisplayName,
		Data:          req.Data,
		DashboardName: req.DashboardName,
		ProjectID:     req.ProjectId,
		UserID:        req.UserId,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.CreateBookmarkResponse{
		DashboardBookmark: bookmarkToPB(bookmark),
	}, nil
}

// RemoveBookmark server removes a bookmark for bookmark id
func (s *Server) RemoveBookmark(ctx context.Context, req *adminv1.RemoveBookmarkRequest) (*adminv1.RemoveBookmarkResponse, error) {
	// Return an empty result if not authenticated.
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() == auth.OwnerTypeAnon {
		return &adminv1.RemoveBookmarkResponse{}, nil
	}

	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, fmt.Errorf("not authenticated as a user")
	}

	err := s.admin.DB.DeleteBookmark(ctx, req.BookmarkId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RemoveBookmarkResponse{}, nil
}

func bookmarkToPB(u *database.Bookmark) *adminv1.DashboardBookmark {
	return &adminv1.DashboardBookmark{
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
