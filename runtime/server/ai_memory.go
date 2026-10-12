package server

import (
	"context"
	"errors"
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
	claims, err := s.checkAIMemoryAccess(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	memories, err := catalog.FindAIMemories(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	res := make([]*runtimev1.AIMemory, len(memories))
	for i, m := range memories {
		res[i] = aiMemoryToPB(m)
	}
	return &runtimev1.ListAIMemoriesResponse{Memories: res}, nil
}

func (s *Server) CreateAIMemory(ctx context.Context, req *runtimev1.CreateAIMemoryRequest) (*runtimev1.CreateAIMemoryResponse, error) {
	claims, err := s.checkAIMemoryAccess(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	content, err := validateAIMemoryInput(req.Content, req.Agent)
	if err != nil {
		return nil, err
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	existing, err := catalog.FindAIMemories(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if len(existing) >= ai.MaxMemories {
		return nil, status.Errorf(codes.ResourceExhausted, "memory limit of %d reached", ai.MaxMemories)
	}

	now := time.Now()
	m := &drivers.AIMemory{
		ID:         uuid.NewString(),
		InstanceID: req.InstanceId,
		OwnerID:    claims.UserID,
		Agent:      req.Agent,
		Content:    content,
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
	claims, err := s.checkAIMemoryAccess(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	content, err := validateAIMemoryInput(req.Content, req.Agent)
	if err != nil {
		return nil, err
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

	m.Agent = req.Agent
	m.Content = content
	err = catalog.UpdateAIMemory(ctx, m)
	if err != nil {
		return nil, err
	}

	return &runtimev1.UpdateAIMemoryResponse{Memory: aiMemoryToPB(m)}, nil
}

func (s *Server) DeleteAIMemory(ctx context.Context, req *runtimev1.DeleteAIMemoryRequest) (*runtimev1.DeleteAIMemoryResponse, error) {
	claims, err := s.checkAIMemoryAccess(ctx, req.InstanceId)
	if err != nil {
		return nil, err
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

func (s *Server) GetAIMemorySettings(ctx context.Context, req *runtimev1.GetAIMemorySettingsRequest) (*runtimev1.GetAIMemorySettingsResponse, error) {
	claims, err := s.checkAIMemoryAccess(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	enabled, err := ai.MemoryEnabledByUser(ctx, catalog, claims.UserID)
	if err != nil {
		return nil, err
	}

	return &runtimev1.GetAIMemorySettingsResponse{Enabled: enabled}, nil
}

func (s *Server) UpdateAIMemorySettings(ctx context.Context, req *runtimev1.UpdateAIMemorySettingsRequest) (*runtimev1.UpdateAIMemorySettingsResponse, error) {
	claims, err := s.checkAIMemoryAccess(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	err = catalog.UpsertAIMemorySettings(ctx, &drivers.AIMemorySettings{
		InstanceID: req.InstanceId,
		OwnerID:    claims.UserID,
		Enabled:    req.Enabled,
	})
	if err != nil {
		return nil, err
	}

	return &runtimev1.UpdateAIMemorySettingsResponse{}, nil
}

// checkAIMemoryAccess checks that the caller may use AI and that memory is available to them.
// It returns a FailedPrecondition error when memory is not available: for anonymous users on Rill Cloud and when the chat_memory feature flag is off.
func (s *Server) checkAIMemoryAccess(ctx context.Context, instanceID string) (*runtime.SecurityClaims, error) {
	claims := auth.GetClaims(ctx, instanceID)
	if !claims.Can(runtime.UseAI) {
		return nil, ErrForbidden
	}
	available, err := ai.MemoryAvailable(ctx, s.runtime, instanceID, claims)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, status.Error(codes.FailedPrecondition, "memory is not available")
	}
	return claims, nil
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

// validateAIMemoryInput validates a user-provided content and agent, and returns the sanitized content.
// The agent must be one of the agents that use memory, or empty for a memory that applies to every agent.
func validateAIMemoryInput(content, agent string) (string, error) {
	content = ai.SanitizeMemoryContent(content)
	if content == "" {
		return "", status.Error(codes.InvalidArgument, "content must not be empty")
	}
	if utf8.RuneCountInString(content) > ai.MaxMemoryContentChars {
		return "", status.Errorf(codes.InvalidArgument, "content must be at most %d characters", ai.MaxMemoryContentChars)
	}
	switch agent {
	case "", ai.AnalystAgentName, ai.DeveloperAgentName:
	default:
		return "", status.Errorf(codes.InvalidArgument, "invalid agent %q", agent)
	}
	return content, nil
}

// aiMemoryToPB converts a drivers.AIMemory to a runtimev1.AIMemory.
func aiMemoryToPB(m *drivers.AIMemory) *runtimev1.AIMemory {
	return &runtimev1.AIMemory{
		Id:                   m.ID,
		Content:              m.Content,
		Agent:                m.Agent,
		SourceConversationId: m.SourceSessionID,
		CreatedOn:            timestamppb.New(m.CreatedOn),
		UpdatedOn:            timestamppb.New(m.UpdatedOn),
	}
}
