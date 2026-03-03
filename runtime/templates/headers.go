package templates

import (
	"fmt"
	"regexp"
	"strings"
)

// sensitiveHeaderPattern matches header keys that carry secret values.
var sensitiveHeaderPattern = regexp.MustCompile(
	`(?i)^(authorization|x-api-key|api-key|token|x-token|x-auth|x-secret|proxy-authorization)$`,
)

// headerKeyCleanupPattern sanitizes a header key into a valid .env variable segment.
// Compiled at package level to avoid per-call regex compilation.
var headerKeyCleanupPattern = regexp.MustCompile(`[^a-z0-9]+`)

// IsSensitiveHeaderKey returns true when a header key likely carries a secret value.
func IsSensitiveHeaderKey(key string) bool {
	return sensitiveHeaderPattern.MatchString(strings.TrimSpace(key))
}

// AuthSchemePrefixes are common HTTP authentication scheme prefixes.
// When a sensitive header value starts with one of these (case-insensitive),
// only the token portion is stored in .env.
var AuthSchemePrefixes = []string{"Bearer ", "Basic ", "Token ", "Bot "}

// SplitAuthSchemePrefix splits a value into scheme prefix and secret if it starts
// with a known auth scheme. Returns ("", "", false) when no prefix matches.
func SplitAuthSchemePrefix(value string) (scheme, secret string, ok bool) {
	for _, prefix := range AuthSchemePrefixes {
		if len(value) > len(prefix) && strings.EqualFold(value[:len(prefix)], prefix) {
			return value[:len(prefix)], value[len(prefix):], true
		}
	}
	return "", "", false
}

// HeaderKeyToEnvSegment sanitizes a header key into a valid .env variable segment.
func HeaderKeyToEnvSegment(key string) string {
	return headerKeyCleanupPattern.ReplaceAllString(strings.ToLower(key), "_")
}

// ResolveHeaderEnvVarName determines the env var name for a header, resolving conflicts.
func ResolveHeaderEnvVarName(connectorName, segment string, existingEnv map[string]bool) string {
	base := fmt.Sprintf("connector.%s.%s", connectorName, segment)
	candidate := base
	for i := 1; existingEnv[candidate]; i++ {
		candidate = fmt.Sprintf("%s_%d", base, i)
	}
	return candidate
}
