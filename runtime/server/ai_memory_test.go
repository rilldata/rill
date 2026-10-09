package server_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAIMemories(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	srv, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	fooCtx := auth.WithClaims(t.Context(), &runtime.SecurityClaims{UserID: "foo", Permissions: []runtime.Permission{runtime.UseAI}})
	barCtx := auth.WithClaims(t.Context(), &runtime.SecurityClaims{UserID: "bar", Permissions: []runtime.Permission{runtime.UseAI}})
	anonCtx := auth.WithClaims(t.Context(), &runtime.SecurityClaims{Permissions: []runtime.Permission{runtime.UseAI}})
	noAICtx := auth.WithClaims(t.Context(), &runtime.SecurityClaims{UserID: "foo"})

	// Anonymous users on Rill Cloud get no memory
	list, err := srv.ListAIMemories(anonCtx, &runtimev1.ListAIMemoriesRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.False(t, list.Enabled)
	require.Empty(t, list.Memories)
	_, err = srv.CreateAIMemory(anonCtx, &runtimev1.CreateAIMemoryRequest{InstanceId: instanceID, Category: "preference", Content: "x"})
	require.Equal(t, codes.FailedPrecondition, status.Code(err))

	// Users without the AI permission are forbidden
	_, err = srv.ListAIMemories(noAICtx, &runtimev1.ListAIMemoriesRequest{InstanceId: instanceID})
	require.Error(t, err)

	// Create
	created, err := srv.CreateAIMemory(fooCtx, &runtimev1.CreateAIMemoryRequest{InstanceId: instanceID, Category: "preference", Content: " Prefers tables "})
	require.NoError(t, err)
	require.Equal(t, "Prefers tables", created.Memory.Content)
	require.Equal(t, "manual", created.Memory.Source)
	require.Equal(t, "active", created.Memory.Status)

	_, err = srv.CreateAIMemory(fooCtx, &runtimev1.CreateAIMemoryRequest{InstanceId: instanceID, Category: "bogus", Content: "x"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	// List is scoped to the owner
	list, err = srv.ListAIMemories(fooCtx, &runtimev1.ListAIMemoriesRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.True(t, list.Enabled)
	require.Len(t, list.Memories, 1)
	list, err = srv.ListAIMemories(barCtx, &runtimev1.ListAIMemoriesRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.Empty(t, list.Memories)

	// Other users cannot update or delete it
	content := "hijacked"
	_, err = srv.UpdateAIMemory(barCtx, &runtimev1.UpdateAIMemoryRequest{InstanceId: instanceID, MemoryId: created.Memory.Id, Content: &content})
	require.Equal(t, codes.NotFound, status.Code(err))
	_, err = srv.DeleteAIMemory(barCtx, &runtimev1.DeleteAIMemoryRequest{InstanceId: instanceID, MemoryId: created.Memory.Id})
	require.Equal(t, codes.NotFound, status.Code(err))

	// Soft delete and restore (undo)
	deleted := "deleted"
	updated, err := srv.UpdateAIMemory(fooCtx, &runtimev1.UpdateAIMemoryRequest{InstanceId: instanceID, MemoryId: created.Memory.Id, Status: &deleted})
	require.NoError(t, err)
	require.Equal(t, "deleted", updated.Memory.Status)
	list, err = srv.ListAIMemories(fooCtx, &runtimev1.ListAIMemoriesRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.Empty(t, list.Memories)
	list, err = srv.ListAIMemories(fooCtx, &runtimev1.ListAIMemoriesRequest{InstanceId: instanceID, IncludeDeleted: true})
	require.NoError(t, err)
	require.Len(t, list.Memories, 1)
	active := "active"
	updated, err = srv.UpdateAIMemory(fooCtx, &runtimev1.UpdateAIMemoryRequest{InstanceId: instanceID, MemoryId: created.Memory.Id, Status: &active})
	require.NoError(t, err)
	require.Equal(t, "active", updated.Memory.Status)

	// Settings
	settings, err := srv.GetAIMemorySettings(fooCtx, &runtimev1.GetAIMemorySettingsRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.True(t, settings.Enabled)
	require.False(t, settings.Paused)
	require.Equal(t, uint32(1), settings.ActiveCount)
	_, err = srv.UpdateAIMemorySettings(fooCtx, &runtimev1.UpdateAIMemorySettingsRequest{InstanceId: instanceID, Paused: true})
	require.NoError(t, err)
	settings, err = srv.GetAIMemorySettings(fooCtx, &runtimev1.GetAIMemorySettingsRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.True(t, settings.Paused)
	settings, err = srv.GetAIMemorySettings(barCtx, &runtimev1.GetAIMemorySettingsRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.False(t, settings.Paused)

	// Delete all
	_, err = srv.DeleteAllAIMemories(fooCtx, &runtimev1.DeleteAllAIMemoriesRequest{InstanceId: instanceID})
	require.NoError(t, err)
	list, err = srv.ListAIMemories(fooCtx, &runtimev1.ListAIMemoriesRequest{InstanceId: instanceID, IncludeDeleted: true})
	require.NoError(t, err)
	require.Empty(t, list.Memories)
}

func TestAIMemoriesFeatureFlagOff(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{"rill.yaml": "features:\n  chat_memory: false\n"},
	})
	srv, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	ctx := auth.WithClaims(t.Context(), &runtime.SecurityClaims{UserID: "foo", Permissions: []runtime.Permission{runtime.UseAI}})
	list, err := srv.ListAIMemories(ctx, &runtimev1.ListAIMemoriesRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.False(t, list.Enabled)
	_, err = srv.CreateAIMemory(ctx, &runtimev1.CreateAIMemoryRequest{InstanceId: instanceID, Category: "preference", Content: "x"})
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
}
