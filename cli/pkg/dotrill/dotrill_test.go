package dotrill

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHomeDirDefault(t *testing.T) {
	d := New("")
	require.NotEmpty(t, d.homeDir)
}

func TestGetSet(t *testing.T) {
	d := New(t.TempDir())

	val, err := d.Get(ConfigFilename, "foo")
	require.NoError(t, err)
	require.Equal(t, "", val)

	err = d.Set(ConfigFilename, "foo", "bar baz")
	require.NoError(t, err)

	val, err = d.Get(ConfigFilename, "foo")
	require.NoError(t, err)
	require.Equal(t, "bar baz", val)

	err = d.Set(ConfigFilename, "hello", "world")
	require.NoError(t, err)

	val, err = d.Get(ConfigFilename, "foo")
	require.NoError(t, err)
	require.Equal(t, "bar baz", val)

	val, err = d.Get(ConfigFilename, "hello")
	require.NoError(t, err)
	require.Equal(t, "world", val)

	err = d.Set(ConfigFilename, "foo", "")
	require.NoError(t, err)

	val, err = d.Get(ConfigFilename, "foo")
	require.NoError(t, err)
	require.Equal(t, "", val)
}

func TestToken(t *testing.T) {
	d := New(t.TempDir())

	creds, err := d.GetAccessToken()
	require.NoError(t, err)
	require.Equal(t, "", creds)

	err = d.SetAccessToken("foo")
	require.NoError(t, err)

	creds, err = d.GetAccessToken()
	require.NoError(t, err)
	require.Equal(t, "foo", creds)

	err = d.SetAccessToken("")
	require.NoError(t, err)

	creds, err = d.GetAccessToken()
	require.NoError(t, err)
	require.Equal(t, "", creds)
}

func TestAnalytics(t *testing.T) {
	d := New(t.TempDir())

	// Test ID gets created
	id1, enabled, err := d.AnalyticsInfo()
	require.NoError(t, err)
	require.True(t, enabled)
	require.Len(t, id1, 36) // UUID string length

	// Test ID is sticky
	id2, enabled, err := d.AnalyticsInfo()
	require.NoError(t, err)
	require.True(t, enabled)
	require.Equal(t, id1, id2)

	// Test it parses analytics_enabled
	d.Set(ConfigFilename, "analytics_enabled", "false")
	id2, enabled, err = d.AnalyticsInfo()
	require.NoError(t, err)
	require.False(t, enabled)
	require.Equal(t, id1, id2)

	// Test it parses analytics_enabled
	d.Set(ConfigFilename, "analytics_enabled", "true")
	id2, enabled, err = d.AnalyticsInfo()
	require.NoError(t, err)
	require.True(t, enabled)
	require.Equal(t, id1, id2)

	// Test it recreates install_id if cleared
	err = d.Set(StateFilename, "install_id", "")
	require.NoError(t, err)
	id3, enabled, err := d.AnalyticsInfo()
	require.NoError(t, err)
	require.True(t, enabled)
	require.NotEqual(t, id1, id3)
	require.Len(t, id3, 36) // UUID string length

	// Test it recreates install_id if state removed
	err = os.Remove(filepath.Join(d.homeDir, ".rill", "state.yaml"))
	require.NoError(t, err)
	id4, enabled, err := d.AnalyticsInfo()
	require.NoError(t, err)
	require.True(t, enabled)
	require.NotEqual(t, id3, id4)
	require.Len(t, id4, 36) // UUID string length
}

func TestAnalyticsMigration(t *testing.T) {
	d := New(t.TempDir())

	// setup resets the homeDir and provides helpers for testing ~/.rill/local.json
	setup := func(t *testing.T) (string, func() bool) {
		d.homeDir = t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(d.homeDir, ".rill"), os.ModePerm))
		oldFilename := filepath.Join(d.homeDir, ".rill", "local.json")
		oldExists := func() bool { _, err := os.Stat(oldFilename); return !os.IsNotExist(err) }
		return oldFilename, oldExists
	}

	t.Run("InstallID only", func(t *testing.T) {
		oldFilename, oldExists := setup(t)

		// Test install ID transfer
		err := os.WriteFile(oldFilename, []byte(`{"installId":"cd29afba-14ff-4cd1-98e6-050a9fb0fee9"}`), 0o644)
		require.NoError(t, err)
		require.True(t, oldExists())

		// Check that it was set correctly
		id, enabled, err := d.AnalyticsInfo()
		require.NoError(t, err)
		require.Equal(t, "cd29afba-14ff-4cd1-98e6-050a9fb0fee9", id)
		require.True(t, enabled)
		require.False(t, oldExists())

		// Repeat, to ensure same is reported second time
		id, enabled, err = d.AnalyticsInfo()
		require.NoError(t, err)
		require.Equal(t, "cd29afba-14ff-4cd1-98e6-050a9fb0fee9", id)
		require.True(t, enabled)
	})

	t.Run("Analytics enabled", func(t *testing.T) {
		oldFilename, oldExists := setup(t)

		err := os.WriteFile(oldFilename, []byte(`{"installId":"cd29afba-14ff-4cd1-98e6-050a9fb0fee9", "analyticsEnabled": true}`), 0o644)
		require.NoError(t, err)
		require.True(t, oldExists())

		id, enabled, err := d.AnalyticsInfo()
		require.NoError(t, err)
		require.Equal(t, "cd29afba-14ff-4cd1-98e6-050a9fb0fee9", id)
		require.True(t, enabled)
		require.False(t, oldExists())
	})

	t.Run("Analytics disabled", func(t *testing.T) {
		oldFilename, oldExists := setup(t)

		err := os.WriteFile(oldFilename, []byte(`{"installId":"cd29afba-14ff-4cd1-98e6-050a9fb0fee9", "analyticsEnabled": false}`), 0o644)
		require.NoError(t, err)
		require.True(t, oldExists())

		id, enabled, err := d.AnalyticsInfo()
		require.NoError(t, err)
		require.Equal(t, "cd29afba-14ff-4cd1-98e6-050a9fb0fee9", id)
		require.False(t, enabled)
		require.False(t, oldExists())
	})

	t.Run("Malformed works", func(t *testing.T) {
		oldFilename, oldExists := setup(t)

		err := os.WriteFile(oldFilename, []byte(`{{"installId":"cd29afba-14ff-4cd1-98e6-050a9fb0fee9", "analyticsEnabled": false}`), 0o644)
		require.NoError(t, err)
		require.True(t, oldExists())

		id, enabled, err := d.AnalyticsInfo()
		require.NoError(t, err)
		require.NotEqual(t, "cd29afba-14ff-4cd1-98e6-050a9fb0fee9", id)
		require.Len(t, id, 36) // UUID string length
		require.True(t, enabled)
		require.True(t, oldExists())

		// Check that second time is persistant, despite malformed local.json
		id2, enabled, err := d.AnalyticsInfo()
		require.NoError(t, err)
		require.Equal(t, id, id2)
		require.True(t, enabled)
		require.True(t, oldExists())
	})

}

func TestRepresentingUserAccessTokenExpiry(t *testing.T) {
	d := New(t.TempDir())

	// Test empty expiry
	expiry, err := d.GetRepresentingUserAccessTokenExpiry()
	require.NoError(t, err)
	require.Zero(t, expiry)

	// Test setting and getting a valid expiry
	now := time.Now().Add(1 * time.Hour)
	err = d.SetRepresentingUserAccessTokenExpiry(now)
	require.NoError(t, err)

	expiry, err = d.GetRepresentingUserAccessTokenExpiry()
	require.NoError(t, err)
	require.True(t, expiry.Equal(now))

	// Test setting and getting zero value expiry
	err = d.SetRepresentingUserAccessTokenExpiry(time.Time{})
	require.NoError(t, err)

	expiry, err = d.GetRepresentingUserAccessTokenExpiry()
	require.NoError(t, err)
	require.Zero(t, expiry)
}
