package ai

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

// This file implements per-user AI memory: short facts and preferences the AI learns about a user in a project.
// Memories are stored in the catalog (see drivers.AIMemory) and injected into agent prompts.
// The AI writes them through the update_memory tool (see memory_update.go); the user manages them through the AI memory RPCs.
//
// Memory is deliberately narrow. It only ever holds what the user said about themselves and their preferences;
// it never holds query results, secrets, or anything the project already states.
// It is scoped to (instance, owner) and is only used by Rill's own chat clients (not by MCP clients or AI reports).

// MemoryFeatureFlag is the feature flag that controls the memory feature. It can be disabled in rill.yaml.
const MemoryFeatureFlag = "chat_memory"

// Memory operations.
const (
	MemoryOpAdd    = "add"
	MemoryOpUpdate = "update"
	MemoryOpDelete = "delete"
	MemoryOpNoop   = "noop"
)

// Limits that bound the size of a user's memory.
const (
	// MaxMemories is the maximum number of memories per (instance, owner).
	MaxMemories = 50
	// MaxMemoryContentChars is the maximum length of a single memory.
	MaxMemoryContentChars = 300
	// maxMemoryPromptChars bounds the rendered memory block in prompts (roughly 3K tokens).
	maxMemoryPromptChars = 12_000
)

// reportUserAgent is the user agent of AI report sessions. Reports render for many recipients, so they never use memory.
const reportUserAgent = "rill/report"

// Patterns that disqualify a candidate memory: secrets and personal identifiers the model should never have proposed.
// They are deliberately narrow; what is worth remembering is left to the prompt, since wider heuristics reject legitimate preferences.
var (
	// Note: "token" only counts as a secret in a credential-like phrase ("access token"), not on its own,
	// because token counts and token spend are ordinary things to analyze in a BI tool.
	memorySecretPattern = regexp.MustCompile(`(?i)\b(password|passwd|secret|bearer|credential|(api|access|auth|refresh|session|id)[ _-]?(key|token))\b`)
	// memoryTokenPattern matches the formats of well-known credentials: OpenAI, GitHub, AWS, Slack, GitLab and Google keys.
	memoryTokenPattern = regexp.MustCompile(`\b(sk-[A-Za-z0-9_-]{8,}|ghp_[A-Za-z0-9]{8,}|github_pat_[A-Za-z0-9_]{8,}|AKIA[0-9A-Z]{16}|xox[abp]-[A-Za-z0-9-]{8,}|glpat-[A-Za-z0-9_-]{8,}|AIza[0-9A-Za-z_-]{20,})`)
	memoryEmailPattern = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)
	// memoryAngleBrackets are stripped on write so a memory can never close the <user_memory> block it is rendered in.
	memoryAngleBrackets = strings.NewReplacer("<", "", ">", "")
)

// MemoryAvailable reports whether memory is available to the given claims in an instance.
// It is false for users without the AI permission, for anonymous users, and when the chat_memory feature flag is off.
func MemoryAvailable(ctx context.Context, rt *runtime.Runtime, instanceID string, claims *runtime.SecurityClaims) (bool, error) {
	if claims == nil || !claims.Can(runtime.UseAI) {
		return false, nil
	}
	// Anonymous users on Rill Cloud (public projects, magic tokens) have no stable identity to key memory on.
	// In Rill Developer, auth is disabled so SkipChecks is true and the empty user ID acts as the single local user.
	if claims.UserID == "" && !claims.SkipChecks {
		return false, nil
	}
	ff, err := rt.FeatureFlags(ctx, instanceID, claims)
	if err != nil {
		return false, err
	}
	return ff[MemoryFeatureFlag], nil
}

// MemoryEnabledByUser reports whether the owner has memory turned on. It is on until the owner turns it off.
func MemoryEnabledByUser(ctx context.Context, catalog drivers.CatalogStore, ownerID string) (bool, error) {
	settings, err := catalog.FindAIMemorySettings(ctx, ownerID)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return true, nil
		}
		return false, err
	}
	return settings.Enabled, nil
}

// loadUserMemories loads the memories to use in a session and reports whether memory is enabled for it.
// Memory is enabled when the client is a Rill chat client (not an MCP client or an AI report),
// memory is available to the owner (see MemoryAvailable), and the owner has not turned it off.
// It returns no memories and false on any error, so a failure never leaves a session with memory half-configured.
func loadUserMemories(ctx context.Context, rt *runtime.Runtime, catalog drivers.CatalogStore, instanceID, userAgent string, claims *runtime.SecurityClaims) ([]*drivers.AIMemory, bool, error) {
	if !strings.HasPrefix(userAgent, "rill") || userAgent == reportUserAgent {
		return nil, false, nil
	}
	available, err := MemoryAvailable(ctx, rt, instanceID, claims)
	if err != nil {
		return nil, false, err
	}
	if !available {
		return nil, false, nil
	}
	enabled, err := MemoryEnabledByUser(ctx, catalog, claims.UserID)
	if err != nil {
		return nil, false, err
	}
	if !enabled {
		return nil, false, nil
	}
	memories, err := catalog.FindAIMemories(ctx, claims.UserID)
	if err != nil {
		return nil, false, err
	}
	return memories, true, nil
}

// memoryIDPrefixLen is the length of the memory ID prefix shown to the model.
// A full UUID is long enough that a model occasionally mistypes one when it calls update_memory;
// the first 8 hex characters are unique among a user's memories (at most MaxMemories) and easy to copy.
const memoryIDPrefixLen = 8

// shortMemoryID returns the prefix of a memory ID that is shown to the model.
func shortMemoryID(id string) string {
	if len(id) > memoryIDPrefixLen {
		return id[:memoryIDPrefixLen]
	}
	return id
}

// findMemoryByID returns the memory with the given ID.
// It accepts the full ID and the prefix shown to the model (see shortMemoryID),
// and finds nothing when the ID is shorter than that prefix or matches more than one memory.
func findMemoryByID(memories []*drivers.AIMemory, id string) (*drivers.AIMemory, bool) {
	if len(id) < memoryIDPrefixLen {
		return nil, false
	}
	var found *drivers.AIMemory
	for _, m := range memories {
		if m.ID == id {
			return m, true
		}
		if strings.HasPrefix(m.ID, id) {
			if found != nil {
				return nil, false
			}
			found = m
		}
	}
	return found, found != nil
}

// renderUserMemories renders the memories that apply to an agent as a list for inclusion in its prompt:
// the memories scoped to it, plus the ones that apply to every agent.
// Each line carries the memory's short ID so the model can reference it in update and delete operations.
func renderUserMemories(memories []*drivers.AIMemory, agent string) string {
	var b strings.Builder
	for _, m := range memories {
		if m.Agent != "" && m.Agent != agent {
			continue
		}
		line := fmt.Sprintf("- %s (id: %s)\n", m.Content, shortMemoryID(m.ID))
		if b.Len()+len(line) > maxMemoryPromptChars {
			break
		}
		b.WriteString(line)
	}
	return strings.TrimRight(b.String(), "\n")
}

// SanitizeMemoryContent prepares content for storage. It runs on every write path, so stored content never needs cleaning when rendered.
// It strips angle brackets, then trims and collapses all whitespace (including newlines) to single spaces,
// so a memory is always a single line that cannot close the <user_memory> block it is rendered in.
func SanitizeMemoryContent(content string) string {
	return strings.Join(strings.Fields(memoryAngleBrackets.Replace(content)), " ")
}

// validateMemoryContent returns a rejection reason if the content must not be remembered, or an empty string if it is acceptable.
func validateMemoryContent(content string) string {
	if content == "" {
		return "empty content"
	}
	if utf8.RuneCountInString(content) > MaxMemoryContentChars {
		return fmt.Sprintf("content exceeds %d characters", MaxMemoryContentChars)
	}
	if memorySecretPattern.MatchString(content) || memoryTokenPattern.MatchString(content) {
		return "content looks like a secret or credential"
	}
	if memoryEmailPattern.MatchString(content) {
		return "content looks like a personal identifier"
	}
	return ""
}

// normalizeMemoryContent normalizes content for duplicate detection.
func normalizeMemoryContent(content string) string {
	content = strings.ToLower(strings.TrimSpace(content))
	content = strings.Join(strings.Fields(content), " ")
	return strings.TrimRight(content, ".!")
}
