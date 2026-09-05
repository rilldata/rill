package databricks

import (
	"context"
	"errors"
	"testing"

	"github.com/databricks/databricks-sql-go/driverctx"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestConn() *connection {
	return &connection{logger: zap.NewNop()}
}

func TestContextWithQueryTags(t *testing.T) {
	t.Run("NoAttributes", func(t *testing.T) {
		// Queries without attributes must be left exactly as they were.
		c := newTestConn()
		ctx := context.Background()
		require.Equal(t, ctx, c.contextWithQueryTags(ctx, nil))
		require.Equal(t, ctx, c.contextWithQueryTags(ctx, map[string]string{}))
		require.Nil(t, driverctx.QueryTagsFromContext(c.contextWithQueryTags(ctx, nil)))
	})

	t.Run("Attributes", func(t *testing.T) {
		c := newTestConn()
		ctx := c.contextWithQueryTags(context.Background(), map[string]string{
			"rillUserEmail": "user@example.com",
			"rillProject":   "my-project",
		})

		require.Equal(t, map[string]string{
			"rillUserEmail": "user@example.com",
			"rillProject":   "my-project",
		}, driverctx.QueryTagsFromContext(ctx))
	})

	t.Run("CopiesAttributes", func(t *testing.T) {
		// The driver reads the map while the query is in flight, so the tags must not
		// alias the caller's map.
		c := newTestConn()
		attrs := map[string]string{"rillUserEmail": "user@example.com"}
		ctx := c.contextWithQueryTags(context.Background(), attrs)

		attrs["rillUserEmail"] = "someone.else@example.com"
		delete(attrs, "rillUserEmail")

		require.Equal(t, map[string]string{"rillUserEmail": "user@example.com"}, driverctx.QueryTagsFromContext(ctx))
	})

	t.Run("SkipsAfterRejection", func(t *testing.T) {
		// Once the workspace has rejected query tags, don't pay for a retry on every
		// subsequent query.
		c := newTestConn()
		c.queryTagsUnsupported.Store(true)

		ctx := context.Background()
		require.Equal(t, ctx, c.contextWithQueryTags(ctx, map[string]string{"rillUserEmail": "user@example.com"}))
	})
}

func TestQueryTagsRejected(t *testing.T) {
	// The error Databricks returns when the workspace has no query tags support.
	unsupported := errors.New("databricks: execution error: failed to execute query: " +
		"[CONFIG_NOT_AVAILABLE.WITHOUT_SUGGESTION] Configuration query_tags is not available.  SQLSTATE: 42K0I")

	t.Run("Unsupported", func(t *testing.T) {
		c := newTestConn()
		require.True(t, c.queryTagsRejected(unsupported))
		require.True(t, c.queryTagsUnsupported.Load())
	})

	t.Run("NilError", func(t *testing.T) {
		c := newTestConn()
		require.False(t, c.queryTagsRejected(nil))
		require.False(t, c.queryTagsUnsupported.Load())
	})

	t.Run("UnrelatedErrorsArePassedThrough", func(t *testing.T) {
		// A genuine query error must not be retried or latched, otherwise real
		// failures would be masked and tags dropped for the rest of the connection.
		c := newTestConn()
		require.False(t, c.queryTagsRejected(errors.New("[TABLE_OR_VIEW_NOT_FOUND] table not found")))
		// A different unavailable configuration is also not ours to handle.
		require.False(t, c.queryTagsRejected(errors.New(
			"[CONFIG_NOT_AVAILABLE.WITHOUT_SUGGESTION] Configuration some_other_conf is not available.")))
		require.False(t, c.queryTagsUnsupported.Load())
	})
}
