package server_test

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAIFeedbackAPIs(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})

	srv, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Create a real conversation owned by another user, so reviewer access is actually exercised.
	ownerClaims := &runtime.SecurityClaims{UserID: "owner-user", SkipChecks: true}
	runner := ai.NewRunner(rt, activity.NewNoopClient())
	session, err := runner.Session(t.Context(), &ai.SessionOptions{
		InstanceID: instanceID,
		Claims:     ownerClaims,
		UserAgent:  "rill/test",
	})
	require.NoError(t, err)
	var listRes *ai.ListMetricsViewsResult
	_, err = session.CallTool(t.Context(), ai.RoleUser, ai.ListMetricsViewsName, &listRes, &ai.ListMetricsViewsArgs{})
	require.NoError(t, err)
	require.NoError(t, session.Flush(t.Context()))
	sessionID := session.CatalogSession().ID

	// Seed feedback items with staggered creation times for pagination.
	catalog, release, err := rt.Catalog(t.Context(), instanceID)
	require.NoError(t, err)
	defer release()
	now := time.Now()
	for i, f := range []*drivers.AIFeedback{
		{ID: "f1", Kind: drivers.AIFeedbackKindReviewRequest, Sentiment: "negative", Comment: "review this", Status: drivers.AIFeedbackStatusOpen},
		{ID: "f2", Kind: drivers.AIFeedbackKindRating, Sentiment: "negative", Comment: "wrong metric", Status: drivers.AIFeedbackStatusOpen},
		{ID: "f3", Kind: drivers.AIFeedbackKindRating, Sentiment: "negative", Comment: "old one", Status: drivers.AIFeedbackStatusAddressed},
	} {
		f.InstanceID = instanceID
		f.SessionID = sessionID
		f.OwnerID = "owner-user"
		f.CreatedOn = now.Add(time.Duration(i-3) * time.Minute)
		f.UpdatedOn = f.CreatedOn
		require.NoError(t, catalog.InsertAIFeedback(t.Context(), f))
	}

	reviewerCtx := auth.WithClaims(t.Context(), &runtime.SecurityClaims{
		UserID:      "reviewer-user",
		Permissions: []runtime.Permission{runtime.UseAI, runtime.ManageAIFeedback},
	})
	nonReviewerCtx := auth.WithClaims(t.Context(), &runtime.SecurityClaims{
		UserID:      "other-user",
		Permissions: []runtime.Permission{runtime.UseAI},
	})

	t.Run("permissions", func(t *testing.T) {
		_, err := srv.ListAIFeedback(nonReviewerCtx, &runtimev1.ListAIFeedbackRequest{InstanceId: instanceID})
		require.ErrorContains(t, err, "action not allowed")
		_, err = srv.GetAIFeedback(nonReviewerCtx, &runtimev1.GetAIFeedbackRequest{InstanceId: instanceID, FeedbackId: "f1"})
		require.ErrorContains(t, err, "action not allowed")
		_, err = srv.UpdateAIFeedbackStatus(nonReviewerCtx, &runtimev1.UpdateAIFeedbackStatusRequest{InstanceId: instanceID, FeedbackId: "f1", Status: "addressed"})
		require.ErrorContains(t, err, "action not allowed")
	})

	t.Run("list with filters and pagination", func(t *testing.T) {
		res, err := srv.ListAIFeedback(reviewerCtx, &runtimev1.ListAIFeedbackRequest{InstanceId: instanceID, Status: "open"})
		require.NoError(t, err)
		require.Len(t, res.Feedback, 2)

		res, err = srv.ListAIFeedback(reviewerCtx, &runtimev1.ListAIFeedbackRequest{InstanceId: instanceID, Kind: "review_request"})
		require.NoError(t, err)
		require.Len(t, res.Feedback, 1)
		require.Equal(t, "f1", res.Feedback[0].Id)

		// Newest first, keyset pagination
		res, err = srv.ListAIFeedback(reviewerCtx, &runtimev1.ListAIFeedbackRequest{InstanceId: instanceID, PageSize: 2})
		require.NoError(t, err)
		require.Len(t, res.Feedback, 2)
		require.Equal(t, "f3", res.Feedback[0].Id)
		require.Equal(t, "f2", res.Feedback[1].Id)
		require.NotEmpty(t, res.NextPageToken)

		res, err = srv.ListAIFeedback(reviewerCtx, &runtimev1.ListAIFeedbackRequest{InstanceId: instanceID, PageSize: 2, PageToken: res.NextPageToken})
		require.NoError(t, err)
		require.Len(t, res.Feedback, 1)
		require.Equal(t, "f1", res.Feedback[0].Id)
	})

	t.Run("reviewer reads transcript, plain conversation access stays denied", func(t *testing.T) {
		res, err := srv.GetAIFeedback(reviewerCtx, &runtimev1.GetAIFeedbackRequest{InstanceId: instanceID, FeedbackId: "f1"})
		require.NoError(t, err)
		require.Equal(t, "f1", res.Feedback.Id)
		require.NotNil(t, res.Conversation)
		require.NotEmpty(t, res.Messages)

		// The same reviewer cannot open the conversation through the regular API:
		// reviewer access is scoped to the feedback review path.
		_, err = srv.GetConversation(reviewerCtx, &runtimev1.GetConversationRequest{InstanceId: instanceID, ConversationId: sessionID})
		require.ErrorContains(t, err, "access denied")
	})

	t.Run("status lifecycle", func(t *testing.T) {
		res, err := srv.UpdateAIFeedbackStatus(reviewerCtx, &runtimev1.UpdateAIFeedbackStatusRequest{InstanceId: instanceID, FeedbackId: "f1", Status: "addressed"})
		require.NoError(t, err)
		require.Equal(t, "addressed", res.Feedback.Status)
		require.Equal(t, "reviewer-user", res.Feedback.ResolvedBy)
		require.NotNil(t, res.Feedback.ResolvedOn)

		// Idempotent: repeating the update returns the current item unchanged
		res, err = srv.UpdateAIFeedbackStatus(reviewerCtx, &runtimev1.UpdateAIFeedbackStatusRequest{InstanceId: instanceID, FeedbackId: "f1", Status: "addressed"})
		require.NoError(t, err)
		require.Equal(t, "addressed", res.Feedback.Status)

		// Reopening clears resolution fields
		res, err = srv.UpdateAIFeedbackStatus(reviewerCtx, &runtimev1.UpdateAIFeedbackStatusRequest{InstanceId: instanceID, FeedbackId: "f1", Status: "open"})
		require.NoError(t, err)
		require.Equal(t, "open", res.Feedback.Status)
		require.Empty(t, res.Feedback.ResolvedBy)
		require.Nil(t, res.Feedback.ResolvedOn)

		_, err = srv.UpdateAIFeedbackStatus(reviewerCtx, &runtimev1.UpdateAIFeedbackStatusRequest{InstanceId: instanceID, FeedbackId: "does-not-exist", Status: "addressed"})
		require.ErrorContains(t, err, "not found")
	})
}
