package templates

import (
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestResolveEnvVarNameWithExplicitName(t *testing.T) {
	prop := &drivers.PropertySpec{Key: "aws_access_key_id", EnvVarName: "AWS_ACCESS_KEY_ID"}
	existing := make(map[string]bool)

	name := ResolveEnvVarName("s3", prop, existing)
	require.Equal(t, "AWS_ACCESS_KEY_ID", name)
}

func TestResolveEnvVarNameFallback(t *testing.T) {
	prop := &drivers.PropertySpec{Key: "password"}
	existing := make(map[string]bool)

	name := ResolveEnvVarName("starrocks", prop, existing)
	require.Equal(t, "STARROCKS_PASSWORD", name)
}

func TestResolveEnvVarNameConflict(t *testing.T) {
	prop := &drivers.PropertySpec{Key: "aws_access_key_id", EnvVarName: "AWS_ACCESS_KEY_ID"}
	existing := map[string]bool{
		"AWS_ACCESS_KEY_ID": true,
	}

	name := ResolveEnvVarName("s3", prop, existing)
	require.Equal(t, "AWS_ACCESS_KEY_ID_1", name)
}

func TestResolveEnvVarNameDoubleConflict(t *testing.T) {
	prop := &drivers.PropertySpec{Key: "aws_access_key_id", EnvVarName: "AWS_ACCESS_KEY_ID"}
	existing := map[string]bool{
		"AWS_ACCESS_KEY_ID":   true,
		"AWS_ACCESS_KEY_ID_1": true,
	}

	name := ResolveEnvVarName("s3", prop, existing)
	require.Equal(t, "AWS_ACCESS_KEY_ID_2", name)
}
