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
	_, err = srv.ListAIMemories(anonCtx, &runtimev1.ListAIMemoriesRequest{InstanceId: instanceID})
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
	_, err = srv.CreateAIMemory(anonCtx, &runtimev1.CreateAIMemoryRequest{InstanceId: instanceID, Content: "x"})
	require.Equal(t, codes.FailedPrecondition, status.Code(err))

	// Users without the AI permission are forbidden
	_, err = srv.ListAIMemories(noAICtx, &runtimev1.ListAIMemoriesRequest{InstanceId: instanceID})
	require.Error(t, err)

	// Create sanitizes the content. A memory the user adds themselves has no source conversation,
	// and applies to every agent unless they pick one.
	created, err := srv.CreateAIMemory(fooCtx, &runtimev1.CreateAIMemoryRequest{InstanceId: instanceID, Content: " Prefers  <tables> "})
	require.NoError(t, err)
	require.Equal(t, "Prefers tables", created.Memory.Content)
	require.Empty(t, created.Memory.Agent)
	require.Empty(t, created.Memory.SourceConversationId)
	scoped, err := srv.CreateAIMemory(fooCtx, &runtimev1.CreateAIMemoryRequest{InstanceId: instanceID, Content: "Names models in snake_case", Agent: "developer_agent"})
	require.NoError(t, err)
	require.Equal(t, "developer_agent", scoped.Memory.Agent)

	_, err = srv.CreateAIMemory(fooCtx, &runtimev1.CreateAIMemoryRequest{InstanceId: instanceID, Content: "   "})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	_, err = srv.CreateAIMemory(fooCtx, &runtimev1.CreateAIMemoryRequest{InstanceId: instanceID, Content: "x", Agent: "router_agent"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	// List is scoped to the owner
	list, err := srv.ListAIMemories(fooCtx, &runtimev1.ListAIMemoriesRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.Len(t, list.Memories, 2)
	list, err = srv.ListAIMemories(barCtx, &runtimev1.ListAIMemoriesRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.Empty(t, list.Memories)

	// Other users cannot update or delete it
	_, err = srv.UpdateAIMemory(barCtx, &runtimev1.UpdateAIMemoryRequest{InstanceId: instanceID, MemoryId: created.Memory.Id, Content: "hijacked"})
	require.Equal(t, codes.NotFound, status.Code(err))
	_, err = srv.DeleteAIMemory(barCtx, &runtimev1.DeleteAIMemoryRequest{InstanceId: instanceID, MemoryId: created.Memory.Id})
	require.Equal(t, codes.NotFound, status.Code(err))

	// Update changes the content and the agent
	updated, err := srv.UpdateAIMemory(fooCtx, &runtimev1.UpdateAIMemoryRequest{InstanceId: instanceID, MemoryId: created.Memory.Id, Content: "Prefers tables, except for time series", Agent: "analyst_agent"})
	require.NoError(t, err)
	require.Equal(t, "Prefers tables, except for time series", updated.Memory.Content)
	require.Equal(t, "analyst_agent", updated.Memory.Agent)
	_, err = srv.UpdateAIMemory(fooCtx, &runtimev1.UpdateAIMemoryRequest{InstanceId: instanceID, MemoryId: created.Memory.Id, Content: "x", Agent: "router_agent"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	// Memory is on until the user turns it off. Turning it off keeps the memories manageable.
	settings, err := srv.GetAIMemorySettings(fooCtx, &runtimev1.GetAIMemorySettingsRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.True(t, settings.Enabled)
	_, err = srv.UpdateAIMemorySettings(fooCtx, &runtimev1.UpdateAIMemorySettingsRequest{InstanceId: instanceID, Enabled: false})
	require.NoError(t, err)
	settings, err = srv.GetAIMemorySettings(fooCtx, &runtimev1.GetAIMemorySettingsRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.False(t, settings.Enabled)
	list, err = srv.ListAIMemories(fooCtx, &runtimev1.ListAIMemoriesRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.Len(t, list.Memories, 2)
	settings, err = srv.GetAIMemorySettings(barCtx, &runtimev1.GetAIMemorySettingsRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.True(t, settings.Enabled)
	_, err = srv.UpdateAIMemorySettings(fooCtx, &runtimev1.UpdateAIMemorySettingsRequest{InstanceId: instanceID, Enabled: true})
	require.NoError(t, err)
	settings, err = srv.GetAIMemorySettings(fooCtx, &runtimev1.GetAIMemorySettingsRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.True(t, settings.Enabled)

	// Delete is permanent
	_, err = srv.DeleteAIMemory(fooCtx, &runtimev1.DeleteAIMemoryRequest{InstanceId: instanceID, MemoryId: created.Memory.Id})
	require.NoError(t, err)
	list, err = srv.ListAIMemories(fooCtx, &runtimev1.ListAIMemoriesRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.Len(t, list.Memories, 1)
	require.Equal(t, scoped.Memory.Id, list.Memories[0].Id)
	_, err = srv.DeleteAIMemory(fooCtx, &runtimev1.DeleteAIMemoryRequest{InstanceId: instanceID, MemoryId: created.Memory.Id})
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestAIMemoriesFeatureFlagOff(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{"rill.yaml": "features:\n  chat_memory: false\n"},
	})
	srv, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	ctx := auth.WithClaims(t.Context(), &runtime.SecurityClaims{UserID: "foo", Permissions: []runtime.Permission{runtime.UseAI}})
	_, err = srv.ListAIMemories(ctx, &runtimev1.ListAIMemoriesRequest{InstanceId: instanceID})
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
	_, err = srv.CreateAIMemory(ctx, &runtimev1.CreateAIMemoryRequest{InstanceId: instanceID, Content: "x"})
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
}
