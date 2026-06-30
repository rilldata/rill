package templates

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveEnvVarNameWithExplicitName(t *testing.T) {
	existing := make(map[string]bool)

	name := ResolveEnvVarNameForKey("s3", "aws_access_key_id", "AWS_ACCESS_KEY_ID", existing)
	require.Equal(t, "AWS_ACCESS_KEY_ID", name)
}

func TestResolveEnvVarNameFallback(t *testing.T) {
	existing := make(map[string]bool)

	name := ResolveEnvVarNameForKey("starrocks", "password", "", existing)
	require.Equal(t, "STARROCKS_PASSWORD", name)
}

func TestResolveEnvVarNameConflict(t *testing.T) {
	existing := map[string]bool{
		"AWS_ACCESS_KEY_ID": true,
	}

	name := ResolveEnvVarNameForKey("s3", "aws_access_key_id", "AWS_ACCESS_KEY_ID", existing)
	require.Equal(t, "AWS_ACCESS_KEY_ID_1", name)
}

func TestResolveEnvVarNameDoubleConflict(t *testing.T) {
	existing := map[string]bool{
		"AWS_ACCESS_KEY_ID":   true,
		"AWS_ACCESS_KEY_ID_1": true,
	}

	name := ResolveEnvVarNameForKey("s3", "aws_access_key_id", "AWS_ACCESS_KEY_ID", existing)
	require.Equal(t, "AWS_ACCESS_KEY_ID_2", name)
}
