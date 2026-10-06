package server

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file implements the RPCs for managing the current user's AI memories.
// Memories are always scoped to the calling user: every read and write is keyed by claims.UserID and the handlers refuse to touch other users' rows.

func (s *Server) ListAIMemories(ctx context.Context, req *runtimev1.ListAIMemoriesRequest) (*runtimev1.ListAIMemoriesResponse, error) {
	claims, enabled, err := s.checkAIMemoryAccess(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return &runtimev1.ListAIMemoriesResponse{Enabled: false}, nil
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	settings, err := catalog.FindAIMemorySettings(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	statuses := []string{drivers.AIMemoryStatusActive}
	if req.IncludeDeleted {
		statuses = append(statuses, drivers.AIMemoryStatusDeleted)
	}
	memories, err := catalog.FindAIMemories(ctx, claims.UserID, statuses)
	if err != nil {
		return nil, err
	}

	res := make([]*runtimev1.AIMemory, len(memories))
	for i, m := range memories {
		res[i] = aiMemoryToPB(m)
	}
	return &runtimev1.ListAIMemoriesResponse{
		Memories: res,
		Enabled:  true,
		Paused:   settings.Paused,
	}, nil
}

func (s *Server) CreateAIMemory(ctx context.Context, req *runtimev1.CreateAIMemoryRequest) (*runtimev1.CreateAIMemoryResponse, error) {
	claims, enabled, err := s.checkAIMemoryAccess(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, status.Error(codes.FailedPrecondition, "memory is not available")
	}

	category, content, err := validateAIMemoryInput(req.Category, req.Content)
	if err != nil {
		return nil, err
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	active, err := catalog.FindAIMemories(ctx, claims.UserID, []string{drivers.AIMemoryStatusActive})
	if err != nil {
		return nil, err
	}
	if len(active) >= ai.MaxActiveMemories {
		return nil, status.Errorf(codes.ResourceExhausted, "memory limit of %d reached", ai.MaxActiveMemories)
	}

	now := time.Now()
	m := &drivers.AIMemory{
		ID:         uuid.NewString(),
		InstanceID: req.InstanceId,
		OwnerID:    claims.UserID,
		Category:   category,
		Content:    content,
		Status:     drivers.AIMemoryStatusActive,
		Source:     drivers.AIMemorySourceManual,
		CreatedOn:  now,
		UpdatedOn:  now,
	}
	err = catalog.InsertAIMemory(ctx, m)
	if err != nil {
		return nil, err
	}

	return &runtimev1.CreateAIMemoryResponse{Memory: aiMemoryToPB(m)}, nil
}

func (s *Server) UpdateAIMemory(ctx context.Context, req *runtimev1.UpdateAIMemoryRequest) (*runtimev1.UpdateAIMemoryResponse, error) {
	claims, enabled, err := s.checkAIMemoryAccess(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, status.Error(codes.FailedPrecondition, "memory is not available")
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	m, err := s.findOwnedAIMemory(ctx, catalog, claims, req.MemoryId)
	if err != nil {
		return nil, err
	}

	if req.Category != nil || req.Content != nil {
		category := m.Category
		if req.Category != nil {
			category = *req.Category
		}
		content := m.Content
		if req.Content != nil {
			content = *req.Content
		}
		category, content, err = validateAIMemoryInput(category, content)
		if err != nil {
			return nil, err
		}
		m.Category = category
		m.Content = content
		m.Source = drivers.AIMemorySourceManual
	}
	if req.Status != nil && *req.Status != m.Status {
		if *req.Status == drivers.AIMemoryStatusActive {
			active, err := catalog.FindAIMemories(ctx, claims.UserID, []string{drivers.AIMemoryStatusActive})
			if err != nil {
				return nil, err
			}
			if len(active) >= ai.MaxActiveMemories {
				return nil, status.Errorf(codes.ResourceExhausted, "memory limit of %d reached", ai.MaxActiveMemories)
			}
		}
		m.Status = *req.Status
	}

	err = catalog.UpdateAIMemory(ctx, m)
	if err != nil {
		return nil, err
	}

	return &runtimev1.UpdateAIMemoryResponse{Memory: aiMemoryToPB(m)}, nil
}

func (s *Server) DeleteAIMemory(ctx context.Context, req *runtimev1.DeleteAIMemoryRequest) (*runtimev1.DeleteAIMemoryResponse, error) {
	claims, enabled, err := s.checkAIMemoryAccess(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, status.Error(codes.FailedPrecondition, "memory is not available")
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	m, err := s.findOwnedAIMemory(ctx, catalog, claims, req.MemoryId)
	if err != nil {
		return nil, err
	}

	err = catalog.DeleteAIMemory(ctx, m.ID)
	if err != nil {
		return nil, err
	}

	return &runtimev1.DeleteAIMemoryResponse{}, nil
}

func (s *Server) DeleteAllAIMemories(ctx context.Context, req *runtimev1.DeleteAllAIMemoriesRequest) (*runtimev1.DeleteAllAIMemoriesResponse, error) {
	claims, enabled, err := s.checkAIMemoryAccess(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, status.Error(codes.FailedPrecondition, "memory is not available")
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	err = catalog.DeleteAIMemories(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	return &runtimev1.DeleteAllAIMemoriesResponse{}, nil
}

func (s *Server) GetAIMemorySettings(ctx context.Context, req *runtimev1.GetAIMemorySettingsRequest) (*runtimev1.GetAIMemorySettingsResponse, error) {
	claims, enabled, err := s.checkAIMemoryAccess(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return &runtimev1.GetAIMemorySettingsResponse{Enabled: false, MaxCount: ai.MaxActiveMemories}, nil
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	settings, err := catalog.FindAIMemorySettings(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	active, err := catalog.FindAIMemories(ctx, claims.UserID, []string{drivers.AIMemoryStatusActive})
	if err != nil {
		return nil, err
	}

	return &runtimev1.GetAIMemorySettingsResponse{
		Enabled:     true,
		Paused:      settings.Paused,
		ActiveCount: uint32(len(active)),
		MaxCount:    ai.MaxActiveMemories,
	}, nil
}

func (s *Server) UpdateAIMemorySettings(ctx context.Context, req *runtimev1.UpdateAIMemorySettingsRequest) (*runtimev1.UpdateAIMemorySettingsResponse, error) {
	claims, enabled, err := s.checkAIMemoryAccess(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, status.Error(codes.FailedPrecondition, "memory is not available")
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	err = catalog.UpsertAIMemorySettings(ctx, &drivers.AIMemorySettings{
		InstanceID: req.InstanceId,
		OwnerID:    claims.UserID,
		Paused:     req.Paused,
	})
	if err != nil {
		return nil, err
	}

	return &runtimev1.UpdateAIMemorySettingsResponse{Paused: req.Paused}, nil
}

// checkAIMemoryAccess checks that the caller may use AI and reports whether memory is available to them.
// Memory is unavailable to anonymous users on Rill Cloud and when the chat_memory feature flag is off.
func (s *Server) checkAIMemoryAccess(ctx context.Context, instanceID string) (*runtime.SecurityClaims, bool, error) {
	claims := auth.GetClaims(ctx, instanceID)
	if !claims.Can(runtime.UseAI) {
		return nil, false, ErrForbidden
	}
	if claims.UserID == "" && !claims.SkipChecks {
		return claims, false, nil
	}
	ff, err := s.runtime.FeatureFlags(ctx, instanceID, claims)
	if err != nil {
		return nil, false, err
	}
	return claims, ff[ai.MemoryFeatureFlag], nil
}

// findOwnedAIMemory loads a memory and verifies it belongs to the caller.
func (s *Server) findOwnedAIMemory(ctx context.Context, catalog drivers.CatalogStore, claims *runtime.SecurityClaims, memoryID string) (*drivers.AIMemory, error) {
	m, err := catalog.FindAIMemory(ctx, memoryID)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "memory not found")
		}
		return nil, err
	}
	if m.OwnerID != claims.UserID {
		return nil, status.Error(codes.NotFound, "memory not found")
	}
	return m, nil
}

// validateAIMemoryInput validates a user-provided category and content.
func validateAIMemoryInput(category, content string) (string, string, error) {
	content = ai.NormalizeMemoryWhitespace(content)
	if content == "" {
		return "", "", status.Error(codes.InvalidArgument, "content must not be empty")
	}
	if utf8.RuneCountInString(content) > ai.MaxMemoryContentChars {
		return "", "", status.Errorf(codes.InvalidArgument, "content must be at most %d characters", ai.MaxMemoryContentChars)
	}
	switch category {
	case ai.MemoryCategoryPreference, ai.MemoryCategoryDefinition, ai.MemoryCategoryContext, ai.MemoryCategoryFeedback:
	case "":
		category = ai.MemoryCategoryContext
	default:
		return "", "", status.Error(codes.InvalidArgument, fmt.Sprintf("invalid category %q", category))
	}
	return category, content, nil
}

// aiMemoryToPB converts a drivers.AIMemory to a runtimev1.AIMemory.
func aiMemoryToPB(m *drivers.AIMemory) *runtimev1.AIMemory {
	return &runtimev1.AIMemory{
		Id:                   m.ID,
		Category:             m.Category,
		Content:              m.Content,
		Status:               m.Status,
		Source:               m.Source,
		SourceConversationId: m.SourceSessionID,
		SourceMessageId:      m.SourceMessageID,
		CreatedOn:            timestamppb.New(m.CreatedOn),
		UpdatedOn:            timestamppb.New(m.UpdatedOn),
	}
}
