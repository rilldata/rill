package runtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateConnectionKeyPreservesNestedJSONTypes(t *testing.T) {
	base := cachedConnectionConfig{instanceID: "instance", name: "connector", driver: "openai"}

	stringConfig := base
	stringConfig.config = map[string]any{"extra_body": map[string]any{"extension": "[END]"}}
	sliceConfig := base
	sliceConfig.config = map[string]any{"extra_body": map[string]any{"extension": []any{"END"}}}
	require.NotEqual(t, generateKey(stringConfig), generateKey(sliceConfig),
		"a string and a JSON array must never reuse one connector handle")

	mapLikeString := base
	mapLikeString.config = map[string]any{"extra_body": "map[a:b]"}
	nestedMap := base
	nestedMap.config = map[string]any{"extra_body": map[string]any{"a": "b"}}
	require.NotEqual(t, generateKey(mapLikeString), generateKey(nestedMap),
		"a string and a JSON object must never reuse one connector handle")
}

func TestGenerateConnectionKeyIsCanonicalAndDoesNotExposeSecrets(t *testing.T) {
	left := cachedConnectionConfig{
		instanceID: "instance", name: "connector", driver: "openai",
		config: map[string]any{
			"api_key":    "super-secret",
			"extra_body": map[string]any{"thinking": map[string]any{"type": "disabled"}, "seed": float64(1)},
		},
	}
	right := cachedConnectionConfig{
		instanceID: "instance", name: "connector", driver: "openai",
		config: map[string]any{
			"extra_body": map[string]any{"seed": float64(1), "thinking": map[string]any{"type": "disabled"}},
			"api_key":    "super-secret",
		},
	}

	leftKey := generateKey(left)
	require.Equal(t, leftKey, generateKey(right), "map insertion order must not change connector identity")
	require.NotContains(t, leftKey, "super-secret", "cache keys must not embed credentials")
}
