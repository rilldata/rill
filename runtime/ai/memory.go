package ai

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

// This file implements per-user AI memory: short facts and preferences the AI learns about a user in a project.
// Memories are stored in the catalog (see drivers.AIMemory), injected into agent prompts, and written either explicitly
// through the update_memory tool or in the background by extract_memories.
//
// Memory is deliberately narrow. It only ever holds what the user said about themselves and their preferences;
// it never holds query results, secrets, or anything the project already states.
// It is scoped to (instance, owner) and is only available to Rill's own chat clients (not to MCP clients or AI reports).

// MemoryFeatureFlag is the feature flag that controls the memory feature. It can be disabled in rill.yaml.
const MemoryFeatureFlag = "chat_memory"

// Memory categories.
const (
	MemoryCategoryPreference = "preference" // how the user likes responses: formatting, charts vs tables, units, default ranges
	MemoryCategoryDefinition = "definition" // how the user interprets business terms
	MemoryCategoryContext    = "context"    // the user's role, team, and which metrics views or dashboards matter to them
	MemoryCategoryFeedback   = "feedback"   // corrections the user has given
)

var memoryCategories = []string{MemoryCategoryPreference, MemoryCategoryDefinition, MemoryCategoryContext, MemoryCategoryFeedback}

// Memory operations.
const (
	MemoryOpAdd    = "add"
	MemoryOpUpdate = "update"
	MemoryOpDelete = "delete"
	MemoryOpNoop   = "noop"
)

// Limits that bound the size and churn of a user's memory.
const (
	// MaxMemories is the maximum number of memories per (instance, owner).
	MaxMemories = 50
	// MaxMemoryContentChars is the maximum length of a single memory.
	MaxMemoryContentChars = 300
	// MaxMemoryOpsPerTurn is the maximum number of memory operations applied in one tool call or extraction.
	MaxMemoryOpsPerTurn = 3
	// maxMemoryPromptChars bounds the rendered memory block in prompts (roughly 3K tokens).
	maxMemoryPromptChars = 12_000
)

// reportUserAgent is the user agent of AI report sessions. Reports render for many recipients, so they never use memory.
const reportUserAgent = "rill/report"

// Patterns that disqualify a candidate memory.
// They catch secrets and personal identifiers the model should never have proposed, and instruction-like text that could act as a prompt injection.
var (
	// Note: "token" only counts as a secret in a credential-like phrase ("access token"), not on its own,
	// because token counts and token spend are ordinary things to analyze in a BI tool.
	memorySecretPattern      = regexp.MustCompile(`(?i)\b(password|passwd|secret|bearer|credential|(api|access|auth|refresh|session|id)[ _-]?(key|token))\b`)
	memoryEmailPattern       = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)
	memoryLongDigitsPattern  = regexp.MustCompile(`\d{12,}|\d{4}[ -]?\d{4}[ -]?\d{4}[ -]?\d{1,7}\b`)
	memoryRenderSanitizer    = strings.NewReplacer("<", "", ">", "")
	memoryInstructionPattern = regexp.MustCompile(`(?i)(ignore|disregard|override)\b.{0,40}\b(instructions|rules|prompt|limit)|system prompt|\bmust (always|never) call\b`)
)

// MemoryOp is one change to the user's memory.
// It is used both as the model's proposal and as the record of what was applied, so the UI can render it.
type MemoryOp struct {
	Op       string `json:"op" jsonschema:"The operation: add, update, delete or noop.,enum=add,enum=update,enum=delete,enum=noop"`
	MemoryID string `json:"memory_id,omitempty" jsonschema:"ID of the memory being updated or deleted."`
	Category string `json:"category,omitempty" jsonschema:"Category of the memory.,enum=preference,enum=definition,enum=context,enum=feedback"`
	Content  string `json:"content,omitempty" jsonschema:"The memory as a short standalone statement about the user."`
	Reason   string `json:"reason,omitempty" jsonschema:"Why the operation was proposed, or why it was rejected."`
}

// memoryEnabled reports whether memory applies to a session with the given user agent and claims.
// It is false for anonymous users, non-Rill clients (MCP), AI reports, and when the feature flag is off.
func memoryEnabled(ctx context.Context, rt *runtime.Runtime, instanceID, userAgent string, claims *runtime.SecurityClaims) (bool, error) {
	if claims == nil || !claims.Can(runtime.UseAI) {
		return false, nil
	}
	// Anonymous users on Rill Cloud (public projects, magic tokens) have no stable identity to key memory on.
	// In Rill Developer, auth is disabled so SkipChecks is true and the empty user ID acts as the single local user.
	if claims.UserID == "" && !claims.SkipChecks {
		return false, nil
	}
	if !strings.HasPrefix(userAgent, "rill") || userAgent == reportUserAgent {
		return false, nil
	}
	ff, err := rt.FeatureFlags(ctx, instanceID, claims)
	if err != nil {
		return false, err
	}
	return ff[MemoryFeatureFlag], nil
}

// loadUserMemories loads the owner's memories and reports whether the owner has memory turned on.
// It returns no memories when the owner has turned memory off.
func loadUserMemories(ctx context.Context, catalog drivers.CatalogStore, ownerID string) ([]*drivers.AIMemory, bool, error) {
	settings, err := catalog.FindAIMemorySettings(ctx, ownerID)
	if err != nil {
		return nil, false, err
	}
	if !settings.Enabled {
		return nil, false, nil
	}
	memories, err := catalog.FindAIMemories(ctx, ownerID)
	if err != nil {
		return nil, false, err
	}
	return memories, true, nil
}

// renderUserMemories renders memories as a list for inclusion in a prompt.
// Each line carries the memory's ID so the model can reference it in update and delete operations.
func renderUserMemories(memories []*drivers.AIMemory) string {
	var b strings.Builder
	for i, m := range memories {
		if i >= MaxMemories {
			break
		}
		// Angle brackets are stripped so a memory cannot close the <user_memory> block it is rendered in.
		line := fmt.Sprintf("- [%s] %s (id: %s)\n", m.Category, memoryRenderSanitizer.Replace(m.Content), m.ID)
		if b.Len()+len(line) > maxMemoryPromptChars {
			break
		}
		b.WriteString(line)
	}
	return strings.TrimRight(b.String(), "\n")
}

// applyMemoryOps validates and applies memory operations for the session's user.
// It is deterministic and never calls an LLM; it is the single write path shared by the update_memory tool and background extraction.
// It returns every op with its final outcome: rejected ops are returned as noops with a reason.
func applyMemoryOps(ctx context.Context, s *Session, ops []MemoryOp, source, sourceMessageID string) ([]MemoryOp, error) {
	if !s.MemoryEnabled() {
		return nil, errors.New("memory is not enabled for this session")
	}
	ownerID := s.Claims().UserID

	catalog, release, err := s.acquireCatalog(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

	existing, err := catalog.FindAIMemories(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]*drivers.AIMemory, len(existing))
	byNormalizedContent := make(map[string]*drivers.AIMemory, len(existing))
	for _, m := range existing {
		byID[m.ID] = m
		byNormalizedContent[normalizeMemoryContent(m.Content)] = m
	}

	now := time.Now()
	result := make([]MemoryOp, 0, len(ops))
	for i, op := range ops {
		op.Content = NormalizeMemoryWhitespace(op.Content)

		// Ops beyond the per-turn budget are rejected rather than dropped, so the caller still sees what was proposed.
		if i >= MaxMemoryOpsPerTurn {
			result = append(result, rejectMemoryOp(op, fmt.Sprintf("exceeded the limit of %d memory operations per turn", MaxMemoryOpsPerTurn)))
			continue
		}

		switch op.Op {
		case MemoryOpAdd, MemoryOpUpdate:
			if reason := validateMemoryContent(op.Content); reason != "" {
				result = append(result, rejectMemoryOp(op, reason))
				continue
			}
			if op.Category != "" && !isValidMemoryCategory(op.Category) {
				op.Category = MemoryCategoryContext
			}
		}

		switch op.Op {
		case MemoryOpAdd:
			if op.Category == "" {
				op.Category = MemoryCategoryContext
			}
			if m, ok := byNormalizedContent[normalizeMemoryContent(op.Content)]; ok {
				op.MemoryID = m.ID
				result = append(result, rejectMemoryOp(op, "already remembered"))
				continue
			}
			if len(byID) >= MaxMemories {
				result = append(result, rejectMemoryOp(op, fmt.Sprintf("memory limit of %d reached; ask the user to remove memories they no longer need", MaxMemories)))
				continue
			}
			m := &drivers.AIMemory{
				ID:              uuid.NewString(),
				InstanceID:      s.InstanceID(),
				OwnerID:         ownerID,
				Category:        op.Category,
				Content:         op.Content,
				Source:          source,
				SourceSessionID: s.ID(),
				SourceMessageID: sourceMessageID,
				CreatedOn:       now,
				UpdatedOn:       now,
			}
			if err := catalog.InsertAIMemory(ctx, m); err != nil {
				return nil, err
			}
			byID[m.ID] = m
			byNormalizedContent[normalizeMemoryContent(m.Content)] = m
			op.MemoryID = m.ID
			result = append(result, op)

		case MemoryOpUpdate:
			m, ok := byID[op.MemoryID]
			if !ok {
				result = append(result, rejectMemoryOp(op, "memory not found"))
				continue
			}
			if op.Category == "" {
				op.Category = m.Category
			}
			if m.Content == op.Content && m.Category == op.Category {
				result = append(result, rejectMemoryOp(op, "no change"))
				continue
			}
			delete(byNormalizedContent, normalizeMemoryContent(m.Content))
			m.Category = op.Category
			m.Content = op.Content
			m.Source = source
			m.SourceSessionID = s.ID()
			m.SourceMessageID = sourceMessageID
			if err := catalog.UpdateAIMemory(ctx, m); err != nil {
				return nil, err
			}
			byNormalizedContent[normalizeMemoryContent(m.Content)] = m
			result = append(result, op)

		case MemoryOpDelete:
			m, ok := byID[op.MemoryID]
			if !ok {
				result = append(result, rejectMemoryOp(op, "memory not found"))
				continue
			}
			if err := catalog.DeleteAIMemory(ctx, m.ID); err != nil {
				return nil, err
			}
			delete(byID, m.ID)
			delete(byNormalizedContent, normalizeMemoryContent(m.Content))
			// Echo what was removed so the notice in the chat can show it.
			op.Category = m.Category
			op.Content = m.Content
			result = append(result, op)

		case MemoryOpNoop:
			result = append(result, op)

		default:
			result = append(result, rejectMemoryOp(op, fmt.Sprintf("unknown operation %q", op.Op)))
		}
	}

	// Refresh the in-memory copy so later turns in this request see the changes.
	memories, enabled, err := loadUserMemories(ctx, catalog, ownerID)
	if err != nil {
		return nil, err
	}
	s.setUserMemories(memories, enabled)

	return result, nil
}

// validateMemoryContent returns a rejection reason if the content must not be remembered, or an empty string if it is acceptable.
func validateMemoryContent(content string) string {
	if content == "" {
		return "empty content"
	}
	if utf8.RuneCountInString(content) > MaxMemoryContentChars {
		return fmt.Sprintf("content exceeds %d characters", MaxMemoryContentChars)
	}
	if memorySecretPattern.MatchString(content) {
		return "content looks like a secret or credential"
	}
	if memoryEmailPattern.MatchString(content) || memoryLongDigitsPattern.MatchString(content) {
		return "content looks like a personal identifier or account number"
	}
	if memoryInstructionPattern.MatchString(content) {
		return "content looks like an instruction rather than a preference"
	}
	return ""
}

func isValidMemoryCategory(category string) bool {
	for _, c := range memoryCategories {
		if c == category {
			return true
		}
	}
	return false
}

// NormalizeMemoryWhitespace trims content and collapses all whitespace (including newlines) to single spaces,
// so a memory is always a single line when rendered into a prompt.
func NormalizeMemoryWhitespace(content string) string {
	return strings.Join(strings.Fields(content), " ")
}

// normalizeMemoryContent normalizes content for duplicate detection.
func normalizeMemoryContent(content string) string {
	content = strings.ToLower(strings.TrimSpace(content))
	content = strings.Join(strings.Fields(content), " ")
	return strings.TrimRight(content, ".!")
}

func rejectMemoryOp(op MemoryOp, reason string) MemoryOp {
	op.Op = MemoryOpNoop
	op.Reason = reason
	return op
}
