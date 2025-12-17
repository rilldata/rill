---
name: AI Feedback System Design
overview: ""
todos: []
---

# AI Feedback System Technical Design

## Overview

Users can upvote/downvote any AI response. Downvoting opens a modal with categorized feedback options. Feedback is stored as tool call messages in the existing `ai_messages` table, making it available for both analytics and LLM context in subsequent conversations.

### Use Cases

1. **Rill team analytics**: Understand how the AI is failing to improve prompts and tools
2. **LLM self-improvement**: AI can see past feedback in conversation context to inform subsequent responses
3. **Project admin insights** (future): Surface questions that failed due to data modeling deficiencies that project developers need to address

## Architecture Decision: Feedback as Tool Call

Instead of a dedicated `ai_feedback` table, feedback is stored as a `user_feedback` tool call/result in the existing `ai_messages` table.

**Advantages:**

- Feedback is part of the conversation - LLM can see it in subsequent calls
- No new table or migration needed
- Leverages existing `CompleteStreaming` endpoint
- Single storage mechanism serves both analytics and LLM context

**Tradeoffs:**

- Analytics queries require filtering by `tool = 'user_feedback'` and parsing JSON content
- Feedback mixed with conversation messages (but easily filterable)

## Deployment Scenarios

### Rill Cloud (Primary)

- Feedback stored in the runtime SQLite database as `ai_messages`
- Runtime databases are exported daily, making feedback available for analytics
- This is the primary use case for business user AI chat

### Rill Developer (Local)

- Feedback stored in the local SQLite database
- **Not exported** - feedback remains on the user's machine
- Acceptable for initial rollout; can add telemetry emission later if needed

## Data Storage

### Message Structure

Feedback is stored in the existing `ai_messages` table with:

| Field | Value |

|-------|-------|

| `role` | `"user"` |

| `type` | `"call"` for the feedback submission, `"result"` for acknowledgment |

| `tool` | `"user_feedback"` |

| `content_type` | `"json"` |

| `content` | JSON payload (see below) |

### Feedback Content (JSON)

```json
{
  "target_message_id": "msg_abc123",
  "sentiment": "negative",
  "categories": ["instruction_ignored", "incorrect_information"],
  "comment": "The chart didn't include the filter I asked for"
}
```

Using key-value strings for flexibility rather than typed enums, since feedback categories are likely to evolve.

### Feedback Categories

| Category Key | Display Name |

|--------------|--------------|

| `instruction_ignored` | Instruction ignored |

| `no_citation_links` | No citation links |

| `being_lazy` | Being lazy |

| `incorrect_information` | Incorrect information |

| `other` | Others |

### Analytics Query Example

```sql
SELECT 
    session_id,
    parent_id,
    created_on,
    json_extract(content, '$.target_message_id') as target_message_id,
    json_extract(content, '$.sentiment') as sentiment,
    json_extract(content, '$.categories') as categories,
    json_extract(content, '$.comment') as comment
FROM ai_messages 
WHERE tool = 'user_feedback' AND type = 'call'
```

## API Design

### Extending `CompleteStreaming`

Add a `UserFeedbackContext` parameter to the existing `CompleteStreamingRequest`:

```protobuf
message CompleteStreamingRequest {
  string instance_id = 1;
  string conversation_id = 2;
  string prompt = 3;
  string agent = 10;
  AnalystAgentContext analyst_agent_context = 11;
  DeveloperAgentContext developer_agent_context = 12;
  
  // NEW: Optional feedback context. If provided, router_agent inserts
  // a user_feedback tool call and exits immediately (no LLM invocation).
  UserFeedbackContext user_feedback_context = 13;
}

message UserFeedbackContext {
  string target_message_id = 1;  // The message being rated
  string sentiment = 2;          // "positive" or "negative"
  repeated string categories = 3; // Only for negative sentiment
  string comment = 4;            // Optional free-text
}
```

**Behavior:**

- If `user_feedback_context` is provided, the `router_agent` adds the feedback as a tool call message and returns immediately
- No LLM is invoked - this is a fast path for feedback submission
- The response includes the `conversation_id` for confirmation

## Backend Implementation

### Files to modify:

1. [`proto/rill/runtime/v1/api.proto`](proto/rill/runtime/v1/api.proto) - Add `UserFeedbackContext` message and field
2. [`runtime/server/chat.go`](runtime/server/chat.go) - Handle feedback context in streaming handler
3. [`runtime/ai/`](runtime/ai/) - Add logic to insert `user_feedback` tool call message

### Implementation Flow

```
Frontend submits feedback
    ↓
CompleteStreaming(user_feedback_context: {...})
    ↓
router_agent checks for feedback context
    ↓
If present: Insert user_feedback tool call → Return immediately
If absent: Normal LLM routing flow
```

## Frontend Implementation

### Components:

1. **Feedback buttons** - Thumbs up/down icons on each AI message bubble
2. **Feedback modal** - Dialog with category checkboxes, comment field, Skip/Submit buttons

### Files to modify:

- `web-common/src/features/chat/` - Add feedback UI components
- Use existing `CompleteStreaming` mutation with new `user_feedback_context` parameter

## Analytics Considerations

### For Use Case 1 (Rill improving AI):

- Daily exports include `ai_messages` table
- Filter by `tool = 'user_feedback'` to extract feedback
- Join with other messages to correlate feedback with AI responses, tool usage, etc.
- Track category distribution over time

### For Use Case 2 (LLM Self-Improvement):

- When building conversation context, `user_feedback` messages are included
- LLM can see patterns like "user marked previous response as incorrect"
- May influence subsequent response quality

### For Use Case 3 (Project admin insights - future):

AI responses can be poor due to **project deficiencies** that only a project admin/developer can fix:

| Deficiency Type | Description | Actionable By |

|-----------------|-------------|---------------|

| Missing `ai_instructions` | Lack of personalized guidance for the AI in the metrics view YAML | Project developer |

| Poor metadata | Missing or inaccurate `measure.description` / `dimension.description` | Project developer |

| Missing data | Required data not modeled or ingested | Project developer |

| Poor data quality | Inconsistent, stale, or incorrect data | Project developer / Data owner |

**Future Enhancement**: Add attribution tracking to distinguish:

- `ai_failure` - Rill's prompt/tool issue (Rill team addresses)
- `project_gap` - Project deficiency (route to project admin)

This could be implemented as:

1. A new `attribution` field in feedback content (manual classification by Rill team initially)
2. Later: AI-assisted classification based on conversation context
3. Surface `project_gap` feedback to project admins via a dashboard or notifications