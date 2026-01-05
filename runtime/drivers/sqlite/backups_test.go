package sqlite

import (
	"fmt"
	"path/filepath"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gocloud.dev/blob/fileblob"
	"google.golang.org/protobuf/proto"
)

func TestBackup(t *testing.T) {
	// File paths for the test
	tmpdir := t.TempDir()
	dbPath := filepath.Join(tmpdir, "data.sqlite")
	storageDir := filepath.Join(tmpdir, "storage")
	bucketDir := filepath.Join(storageDir, "bucket")

	// Create local bucket
	bucket, err := fileblob.OpenBucket(bucketDir, &fileblob.Options{CreateDir: true})
	require.NoError(t, err)

	// Create sqlite handle
	cfg := map[string]any{
		"dsn": dbPath,
		"id":  "test-backup",
	}
	h, err := driver{}.Open("", cfg, storage.MustNew(storageDir, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer h.Close()
	err = h.Migrate(t.Context())
	require.NoError(t, err)

	// Insert some random data into every table
	registry, ok := h.AsRegistry()
	require.True(t, ok)
	catalog, ok := h.AsCatalogStore("x")
	require.True(t, ok)
	require.NoError(t, registry.CreateInstance(t.Context(), &drivers.Instance{
		ID: "a",
	}))
	v, err := catalog.NextControllerVersion(t.Context())
	require.NoError(t, err)
	require.NoError(t, catalog.CreateResource(t.Context(), v, drivers.Resource{
		Kind: "a",
		Name: "b",
		Data: must(proto.Marshal(&runtimev1.Resource{Meta: &runtimev1.ResourceMeta{Name: &runtimev1.ResourceName{Kind: "a", Name: "b"}}})),
	}))
	require.NoError(t, catalog.InsertModelPartition(t.Context(), "model1", drivers.ModelPartition{
		Key:      "a",
		DataJSON: []byte(`{"b":"c"}`),
	}))
	require.NoError(t, catalog.UpsertInstanceHealth(t.Context(), &drivers.InstanceHealth{
		InstanceID: "a",
		HealthJSON: []byte(`{"b":"c"}`),
	}))
	require.NoError(t, catalog.InsertAISession(t.Context(), &drivers.AISession{
		ID: "a",
	}))
	require.NoError(t, catalog.InsertAIMessage(t.Context(), &drivers.AIMessage{
		ID:        "a",
		SessionID: "a",
	}))

	// Run a backup
	err = h.(*connection).backup(t.Context(), bucket)
	require.NoError(t, err)

	// Check it created the expected files
	expected := []string{
		"snapshot.db",
		"instances.parquet",
		"instance_health.parquet",
		"catalog.parquet",
		"model_partitions.parquet",
		"ai_sessions.parquet",
		"ai_messages.parquet",
	}
	for _, filename := range expected {
		attr, err := bucket.Attributes(t.Context(), filename)
		require.NoError(t, err, "expected backup file %q to exist", filename)
		require.Greater(t, attr.Size, int64(0), "expected backup file %q to be non-empty", filename)
	}
}

func TestDBFilePath(t *testing.T) {
	tmpDir, _ := filepath.EvalSymlinks(t.TempDir())
	cases := []struct {
		dsn      string
		expected string
	}{
		{":memory:", ""},
		{"file:rill?mode=memory&cache=shared", ""},
		{"file::memory:?cache=shared", ""},
		{filepath.Join(tmpDir, "data.sqlite"), filepath.Join(tmpDir, "data.sqlite")},
		{"file:" + filepath.Join(tmpDir, "data.sqlite"), filepath.Join(tmpDir, "data.sqlite")},
	}
	for idx, tc := range cases {
		t.Run(fmt.Sprintf("case-%d", idx), func(t *testing.T) {
			cfg := map[string]any{"dsn": tc.dsn}
			h, err := driver{}.Open("", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
			require.NoError(t, err)
			defer h.Close()

			dbPath, err := h.(*connection).dbFilePath(t.Context())
			require.NoError(t, err)
			require.Equal(t, tc.expected, dbPath)
		})
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
