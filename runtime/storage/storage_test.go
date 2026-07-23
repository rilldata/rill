package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// These cases pin the constructor's ownership of its scratch directory and its error paths.
	t.Run("initializes local client", func(t *testing.T) {
		// A successful client must have a private scratch directory without eagerly creating its data directory.
		dataDir := filepath.Join(t.TempDir(), "data")
		client, err := New(dataDir, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.RemoveAll(client.tempDirPath) })

		require.Equal(t, dataDir, client.dataDirPath)
		require.Nil(t, client.bucketConfig)
		require.Nil(t, client.prefixes)
		require.DirExists(t, client.tempDirPath)
		require.NoDirExists(t, dataDir)
	})

	t.Run("decodes bucket configuration", func(t *testing.T) {
		// Decoding is tested without opening the bucket so the constructor remains a local-only operation.
		client, err := New(t.TempDir(), map[string]any{
			"bucket":                              "test-bucket",
			"google_application_credentials_json": "test-credentials",
		})
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.RemoveAll(client.tempDirPath) })

		require.Equal(t, &gcsBucketConfig{
			Bucket:                           "test-bucket",
			GoogleApplicationCredentialsJSON: "test-credentials",
		}, client.bucketConfig)
	})

	t.Run("propagates temp directory creation failure", func(t *testing.T) {
		// Pointing TMPDIR at a regular file deterministically makes MkdirTemp fail before client construction.
		tmpFile := filepath.Join(t.TempDir(), "not-a-directory")
		require.NoError(t, os.WriteFile(tmpFile, []byte("fixture"), 0o600))
		t.Setenv("TMPDIR", tmpFile)

		client, err := New(filepath.Join(t.TempDir(), "data"), nil)
		require.Error(t, err)
		require.Nil(t, client)
	})

	t.Run("cleans temp directory after configuration failure", func(t *testing.T) {
		// New owns the scratch directory, so a later decode error must not leak it.
		tmpParent := filepath.Join(t.TempDir(), "scratch")
		require.NoError(t, os.Mkdir(tmpParent, 0o700))
		t.Setenv("TMPDIR", tmpParent)

		client, err := New(filepath.Join(t.TempDir(), "data"), map[string]any{
			"bucket": make(chan struct{}),
		})
		require.Error(t, err)
		require.Nil(t, client)

		entries, readErr := os.ReadDir(tmpParent)
		require.NoError(t, readErr)
		require.Empty(t, entries, "constructor-owned scratch directory leaked after decode failure")
	})
}

func TestClient_WithPrefix(t *testing.T) {
	// Derived clients must compose prefixes without mutating the clients they were derived from.
	root := t.TempDir()
	bucketConfig := &gcsBucketConfig{Bucket: "test-bucket"}
	client := &Client{
		dataDirPath:  filepath.Join(root, "data"),
		tempDirPath:  filepath.Join(root, "temp"),
		bucketConfig: bucketConfig,
	}

	tenantClient := client.WithPrefix("tenant")
	projectClient := tenantClient.WithPrefix("project", "run")

	require.Nil(t, client.prefixes)
	require.Equal(t, []string{"tenant"}, tenantClient.prefixes)
	require.Equal(t, []string{"tenant", "project", "run"}, projectClient.prefixes)
	require.Same(t, bucketConfig, projectClient.bucketConfig)

	dataDir, err := projectClient.DataDir("artifact")
	require.NoError(t, err)
	tempDir, err := projectClient.TempDir("artifact")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(root, "data", "tenant", "project", "run", "artifact"), dataDir)
	require.Equal(t, filepath.Join(root, "temp", "tenant", "project", "run", "artifact"), tempDir)
}

func TestClient_RemovePrefixContainment(t *testing.T) {
	// Each case proves that removal is confined to the client's two local roots.
	t.Run("empty prefix removes only storage roots", func(t *testing.T) {
		// Separate parents make the sibling sentinels meaningful even when both roots are intentionally removed.
		fixture := newRemovePrefixFixture(t)

		err := fixture.client.RemovePrefix(context.Background())
		require.NoError(t, err)
		require.NoDirExists(t, fixture.dataRoot)
		require.NoDirExists(t, fixture.tempRoot)
		require.FileExists(t, fixture.dataOutsideSentinel)
		require.FileExists(t, fixture.tempOutsideSentinel)
	})

	t.Run("normal prefix removes only matching subtrees", func(t *testing.T) {
		// Target and sibling files distinguish the requested prefix from the rest of each root.
		fixture := newRemovePrefixFixture(t)
		dataTarget := writeFixtureFile(t, fixture.dataRoot, "target", "remove-me")
		tempTarget := writeFixtureFile(t, fixture.tempRoot, "target", "remove-me")
		dataSibling := writeFixtureFile(t, fixture.dataRoot, "keep", "sentinel")
		tempSibling := writeFixtureFile(t, fixture.tempRoot, "keep", "sentinel")

		err := fixture.client.RemovePrefix(context.Background(), "target")
		require.NoError(t, err)
		require.NoFileExists(t, dataTarget)
		require.NoFileExists(t, tempTarget)
		require.FileExists(t, dataSibling)
		require.FileExists(t, tempSibling)
		require.FileExists(t, fixture.dataOutsideSentinel)
		require.FileExists(t, fixture.tempOutsideSentinel)
	})

	traversalCases := []struct {
		name   string
		prefix []string
	}{
		{name: "parent traversal", prefix: []string{".."}},
		{name: "nested parent traversal", prefix: []string{"nested", "..", ".."}},
	}
	for _, tt := range traversalCases {
		t.Run(tt.name, func(t *testing.T) {
			// Roots live one level below sentinels so an unvalidated traversal would visibly delete outside data.
			fixture := newRemovePrefixFixture(t)

			err := fixture.client.RemovePrefix(context.Background(), tt.prefix...)
			assert.Error(t, err, "traversal prefix should be rejected")
			assert.FileExists(t, fixture.dataRootSentinel)
			assert.FileExists(t, fixture.tempRootSentinel)
			assert.FileExists(t, fixture.dataOutsideSentinel)
			assert.FileExists(t, fixture.tempOutsideSentinel)
		})
	}
}

func TestClient_RemovePrefixRejectsPrefixedClient(t *testing.T) {
	// A derived client must not reinterpret a removal prefix relative to its hidden base prefix.
	fixture := newRemovePrefixFixture(t)
	dataSentinel := writeFixtureFile(t, fixture.dataRoot, "tenant", "child", "data")
	tempSentinel := writeFixtureFile(t, fixture.tempRoot, "tenant", "child", "temp")

	err := fixture.client.WithPrefix("tenant").RemovePrefix(context.Background(), "child")
	require.EqualError(t, err, "storage: RemovePrefix is not supported for prefixed client")
	require.FileExists(t, dataSentinel)
	require.FileExists(t, tempSentinel)
	require.FileExists(t, fixture.dataOutsideSentinel)
	require.FileExists(t, fixture.tempOutsideSentinel)
}

func TestClient_RemovePrefixLocalFailureCleanup(t *testing.T) {
	// Removal must report either local failure while still attempting cleanup of the other root.
	t.Run("data removal failure still cleans temp prefix", func(t *testing.T) {
		// A regular file used as dataDir makes removal below it fail with ENOTDIR on every local filesystem.
		root := t.TempDir()
		dataFile := filepath.Join(root, "data-file")
		require.NoError(t, os.WriteFile(dataFile, []byte("fixture"), 0o600))
		tempRoot := filepath.Join(root, "temp")
		tempTarget := writeFixtureFile(t, tempRoot, "target", "remove-me")
		client := &Client{dataDirPath: dataFile, tempDirPath: tempRoot}

		err := client.RemovePrefix(context.Background(), "target")
		require.Error(t, err)
		require.FileExists(t, dataFile)
		require.NoFileExists(t, tempTarget)
	})

	t.Run("temp removal failure follows data cleanup", func(t *testing.T) {
		// A regular file used as tempDir forces the second removal to fail after data cleanup succeeds.
		root := t.TempDir()
		dataRoot := filepath.Join(root, "data")
		dataTarget := writeFixtureFile(t, dataRoot, "target", "remove-me")
		tempFile := filepath.Join(root, "temp-file")
		require.NoError(t, os.WriteFile(tempFile, []byte("fixture"), 0o600))
		client := &Client{dataDirPath: dataRoot, tempDirPath: tempFile}

		err := client.RemovePrefix(context.Background(), "target")
		require.Error(t, err)
		require.NoFileExists(t, dataTarget)
		require.FileExists(t, tempFile)
	})
}

func TestClient_RemovePrefixBucketFailureAfterLocalCleanup(t *testing.T) {
	// Bucket setup failures must be returned after deterministic local cleanup has completed.
	fixture := newRemovePrefixFixture(t)
	dataTarget := writeFixtureFile(t, fixture.dataRoot, "target", "remove-me")
	tempTarget := writeFixtureFile(t, fixture.tempRoot, "target", "remove-me")
	fixture.client.bucketConfig = &gcsBucketConfig{
		Bucket:                           "unused-test-bucket",
		GoogleApplicationCredentialsJSON: "{invalid-json",
	}

	err := fixture.client.RemovePrefix(context.Background(), "target")
	require.ErrorContains(t, err, "could not create GCP client")
	require.NoFileExists(t, dataTarget)
	require.NoFileExists(t, tempTarget)
	require.FileExists(t, fixture.dataOutsideSentinel)
	require.FileExists(t, fixture.tempOutsideSentinel)
}

type removePrefixFixture struct {
	client              *Client
	dataRoot            string
	tempRoot            string
	dataRootSentinel    string
	tempRootSentinel    string
	dataOutsideSentinel string
	tempOutsideSentinel string
}

func newRemovePrefixFixture(t *testing.T) *removePrefixFixture {
	t.Helper()
	// Independent parents prevent one traversal deletion from masking whether cleanup escaped the other root.
	sandbox := t.TempDir()
	dataParent := filepath.Join(sandbox, "data-parent")
	tempParent := filepath.Join(sandbox, "temp-parent")
	dataRoot := filepath.Join(dataParent, "root")
	tempRoot := filepath.Join(tempParent, "root")

	return &removePrefixFixture{
		client:              &Client{dataDirPath: dataRoot, tempDirPath: tempRoot},
		dataRoot:            dataRoot,
		tempRoot:            tempRoot,
		dataRootSentinel:    writeFixtureFile(t, dataRoot, "root-sentinel"),
		tempRootSentinel:    writeFixtureFile(t, tempRoot, "root-sentinel"),
		dataOutsideSentinel: writeFixtureFile(t, dataParent, "outside-sentinel"),
		tempOutsideSentinel: writeFixtureFile(t, tempParent, "outside-sentinel"),
	}
}

func writeFixtureFile(t *testing.T, elem ...string) string {
	t.Helper()
	path := filepath.Join(elem...)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte("sentinel"), 0o600))
	return path
}

func TestClient_DataDir(t *testing.T) {
	tempDir := t.TempDir()
	client := &Client{
		dataDirPath: tempDir,
	}

	client = client.WithPrefix("testprefix")

	tests := []struct {
		name string
		elem []string
	}{
		{
			name: "create single directory",
			elem: []string{"testdir"},
		},
		{
			name: "create nested directories",
			elem: []string{"testdir", "nested"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.DataDir(tt.elem...)
			require.NoError(t, err)
			if _, err := os.Stat(got); os.IsNotExist(err) {
				t.Errorf("Client.DataDir() path = %v, directory does not exist", got)
			}
			require.Equal(t, filepath.Join(append([]string{tempDir, "testprefix"}, tt.elem...)...), got)
		})
	}
}

func TestClient_TempDir(t *testing.T) {
	tempDir := os.TempDir()
	client := &Client{
		dataDirPath: tempDir,
	}
	client = client.WithPrefix("testprefix", "testtempdir")

	tests := []struct {
		name string
		elem []string
	}{
		{
			name: "create single temp directory",
			elem: []string{"testtempdir"},
		},
		{
			name: "create nested temp directories",
			elem: []string{"testtempdir", "nested"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.TempDir(tt.elem...)
			require.NoError(t, err)
			if _, err := os.Stat(got); os.IsNotExist(err) {
				t.Errorf("Client.TempDir() path = %v, directory does not exist", got)
			}
			require.Equal(t, filepath.Join(append([]string{client.tempDirPath, "testprefix", "testtempdir"}, tt.elem...)...), got)
		})
	}
}

func TestClient_RandomTempDir(t *testing.T) {
	tempDir := t.TempDir()
	client := &Client{
		dataDirPath: tempDir,
	}

	tests := []struct {
		name    string
		pattern string
		elem    []string
	}{
		{
			name:    "create single random temp directory",
			pattern: "testtempdir-*",
			elem:    []string{"random"},
		},
		{
			name:    "create nested random temp directories",
			pattern: "testtempdir-*",
			elem:    []string{"random", "nested"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.RandomTempDir(tt.pattern, tt.elem...)
			require.NoError(t, err)
			if _, err := os.Stat(got); os.IsNotExist(err) {
				t.Errorf("Client.RandomTempDir() path = %v, directory does not exist", got)
			}
			require.Equal(t, filepath.Join(append([]string{client.tempDirPath}, tt.elem...)...), filepath.Dir(got))
		})
	}
}
