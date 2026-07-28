package admin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserFilesDirSegment(t *testing.T) {
	// Human-readable display name plus the full user ID.
	require.Equal(t,
		"jane-doe-018f3a2b-7c4d-4e5f-9a1b-2c3d4e5f6a7b",
		UserFilesDirSegment("018f3a2b-7c4d-4e5f-9a1b-2c3d4e5f6a7b", "Jane Doe", "jane.doe@acme.com"))

	// Empty display name falls back to the email local part.
	require.Equal(t,
		"jane-doe-user-1",
		UserFilesDirSegment("user-1", "", "jane.doe@acme.com"))

	// Non-sanitizable display name and email fall back to a generic prefix.
	require.Equal(t, "user-user-2", UserFilesDirSegment("user-2", "🚀", "@example.com"))

	// Same name, different user IDs => different segments (collision-free).
	a := UserFilesDirSegment("user-1", "John", "john@a.com")
	b := UserFilesDirSegment("user-2", "John", "john@b.com")
	require.NotEqual(t, a, b)
	require.Equal(t, "john-user-1", a)
	require.Equal(t, "john-user-2", b)

	// Odd characters in the display name are sanitized (non-ascii characters are dropped).
	require.Equal(t, "ana-mara-oconnor-u1", UserFilesDirSegment("u1", "Ana_María  O'Connor", ""))

	// Empty owner maps to the shared segment.
	require.Equal(t, UserFilesSharedSegment, UserFilesDirSegment("", "Jane Doe", "jane@acme.com"))
}

func TestUserFilesPath(t *testing.T) {
	require.Equal(t, "user_files/jane-doe-u1/alerts/my-alert.yaml",
		UserFilesPath("jane-doe-u1", UserFilesSubdirAlert, "my-alert"))
}

func TestUserFilesRewriteLegacyPath(t *testing.T) {
	subdir, name, ok := UserFilesRewriteLegacyPath("alerts/my-alert-1a2b3c4d.yaml")
	require.True(t, ok)
	require.Equal(t, UserFilesSubdirAlert, subdir)
	require.Equal(t, "my-alert-1a2b3c4d", name)

	subdir, name, ok = UserFilesRewriteLegacyPath("reports/weekly-9f8e7d6c.yaml")
	require.True(t, ok)
	require.Equal(t, UserFilesSubdirReport, subdir)
	require.Equal(t, "weekly-9f8e7d6c", name)

	// Legacy personal files map to the canvas subdir.
	subdir, name, ok = UserFilesRewriteLegacyPath("personal/my-canvas-5b3f7e1a.yaml")
	require.True(t, ok)
	require.Equal(t, UserFilesSubdirCanvas, subdir)
	require.Equal(t, "my-canvas-5b3f7e1a", name)

	// Paths already in the new layout, nested paths, and unknown dirs are not rewritten.
	for _, p := range []string{
		"user_files/jane-doe-u1/alerts/my-alert.yaml",
		"alerts/nested/my-alert.yaml",
		"models/orders.yaml",
		"alerts/my-alert.sql",
		"rill.yaml",
	} {
		_, _, ok := UserFilesRewriteLegacyPath(p)
		require.False(t, ok, "expected %q not to be rewritten", p)
	}
}
