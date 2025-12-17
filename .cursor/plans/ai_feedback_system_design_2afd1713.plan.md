---
name: AI Feedback System Design
overview: ""
todos: []
---

# AI Feedback System Technical Design

## Overview

Users can upvote/downvote any AI response. Downvoting opens a modal with categorized feedback options. Feedback is stored as tool call messages in the existing `ai_messages` table, making it available for both analytics and LLM context in subsequent conversations.

For negative feedback, the system asynchronously invokes an LLM to predict attribution (`rill` vs `project`), enabling automatic triage without blocking the user.

### Use Cases

1. **Rill team analytics**: Understand how the AI is failing to improve prompts and tools
2. **LLM self-improvement**: AI can see past feedback in conversation context to inform subsequent responses
3. **Project admin insights**: Surface questions that failed due to data modeling deficiencies that project developers need to address

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

| `type` | `"call"` for the feedback submission, `"result"` for the response |

| `tool` | `"user_feedback"` |

| `content_type` | `"json"` |

| `content` | JSON payload (see below) |

### Feedback Call Content (JSON)

```json
{
  "target_message_id": "msg_abc123",
  "sentiment": "negative",
  "categories": ["instruction_ignored", "incorrect_information"],
  "comment": "The chart didn't include the filter I asked for"
}
```

Using key-value strings for flexibility rather than typed enums, since feedback categories are likely to evolve.

### Feedback Result Content (JSON)

**For positive feedback** - simple acknowledgment:

```json
{
  "status": "recorded"
}
```

**For negative feedback** - includes AI-predicted attribution (populated async):

```json
{
  "status": "recorded",
  "predicted_attribution": "project",
  "attribution_reasoning": "User asked about 'customer churn rate' but no churn-related measure exists in the metrics view.",
  "suggested_action": "Add a measure for 'customer_churn_rate' with a clear description explaining how churn is calculated."
}
```

### Feedback Categories

| Category Key | Display Name |

|--------------|--------------|

| `instruction_ignored` | Instruction ignored |

| `no_citation_links` | No citation links |

| `being_lazy` | Being lazy |

| `incorrect_information` | Incorrect information |

| `other` | Others |

### Attribution Types

| Attribution | Description | Actionable By |

|-------------|-------------|---------------|

| `rill` | AI made an error in reasoning, tool usage, or response generation | Rill team |

| `project` | Missing or insufficient project data/metadata | Project developer |

### Analytics Query Example

```sql
SELECT 
    session_id,
    created_on,
    json_extract(content, '$.target_message_id') as target_message_id,
    json_extract(content, '$.sentiment') as sentiment,
    json_extract(content, '$.categories') as categories,
    json_extract(content, '$.comment') as comment
FROM ai_messages 
WHERE tool = 'user_feedback' AND type = 'call'
```
```sql
-- Get attribution predictions for negative feedback
SELECT 
    m1.session_id,
    m1.created_on,
    json_extract(m2.content, '$.predicted_attribution') as attribution,
    json_extract(m2.content, '$.suggested_action') as suggested_action
FROM ai_messages m1
JOIN ai_messages m2 ON m1.parent_id = m2.parent_id 
    AND m2.type = 'result' AND m2.tool = 'user_feedback'
WHERE m1.tool = 'user_feedback' AND m1.type = 'call'
    AND json_extract(m1.content, '$.sentiment') = 'negative'
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
  
  // NEW: Optional feedback context. If provided, records feedback
  // and returns immediately. Attribution prediction runs async.
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

- If `user_feedback_context` is provided: Record feedback call, return immediately (user not blocked)
- For negative sentiment: Spawn goroutine to run attribution prediction, store result when complete

## Async Attribution Flow

To avoid blocking the user, attribution prediction runs asynchronously:

```
Frontend submits negative feedback
    ↓
CompleteStreaming(user_feedback_context: {...})
    ↓
Record user_feedback tool "call" message
    ↓
Return "recorded" acknowledgment to user immediately
    ↓
Spawn goroutine:
    → Invoke LLM with attribution prompt
    → Store "result" message with prediction when complete
```

**Implementation:**

```go
// Record feedback call immediately
s.EmitMessage(ctx, feedbackCallMessage)

// For positive feedback, just record simple result and return
if sentiment == "positive" {
    s.EmitMessage(ctx, simpleResultMessage)
    return
}

// For negative feedback, return immediately, run attribution async
go func() {
    bgCtx := context.Background() // Detached context
    result := predictAttribution(bgCtx, ...)
    s.EmitMessage(bgCtx, resultMessageWithAttribution)
}()
```

## AI Attribution Prediction

For negative feedback, the system invokes an LLM to analyze the feedback and predict attribution.

### Structured Output

The attribution response uses **JSON Schema-constrained output** to guarantee valid structure:

```go
// Define typed result struct
type FeedbackAttributionResult struct {
    PredictedAttribution string  `json:"predicted_attribution" jsonschema:"enum=rill,project"`
    AttributionReasoning string  `json:"attribution_reasoning"`
    SuggestedAction      *string `json:"suggested_action,omitempty"`
}

// Infer schema and pass to Complete
outputSchema, _ := jsonschema.For[*FeedbackAttributionResult]()

res, err := llm.Complete(ctx, &drivers.CompleteOptions{
    Messages:     messages,
    OutputSchema: outputSchema,  // LLM constrained to this schema
})
```

This leverages OpenAI's structured output feature - the model is guaranteed to return JSON matching the schema. The `enum=rill,project` constraint ensures `predicted_attribution` can only be one of those two values.

### Attribution Prompt Context

The LLM receives:

1. The original user question/prompt
2. The AI response that was downvoted
3. The feedback categories and comment from the user
4. Project metadata (available measures, dimensions, ai_instructions)

### Attribution Prompt

```
Analyze this user feedback on an AI response and determine the root cause.

User's original question: {original_prompt}
AI's response: {ai_response}
User's feedback categories: {categories}
User's comment: {comment}

Available project metadata:
- Measures: {measures}
- Dimensions: {dimensions}
- AI Instructions: {ai_instructions}

Determine if this failure was caused by:
1. "rill" - The AI made an error in reasoning, misunderstood the question, used tools incorrectly, or generated an incorrect response
2. "project" - The data or metadata needed to answer correctly is missing or insufficient (e.g., missing measures, poor descriptions, no AI instructions)

For "project" attribution, provide a specific suggested_action the developer can take. For "rill", set suggested_action to null.
```

### Project Gap Types

When attribution is `project`, the reasoning should identify which deficiency applies:

| Deficiency Type | Description |

|-----------------|-------------|

| Missing `ai_instructions` | Lack of personalized guidance for the AI in the metrics view YAML |

| Poor metadata | Missing or inaccurate `measure.description` / `dimension.description` |

| Missing data | Required data not modeled or ingested |

| Poor data quality | Inconsistent, stale, or incorrect data |

## Backend Implementation

### Files to modify:

1. [`proto/rill/runtime/v1/api.proto`](proto/rill/runtime/v1/api.proto) - Add `UserFeedbackContext` message and field
2. [`runtime/server/chat.go`](runtime/server/chat.go) - Handle feedback context in streaming handler
3. [`runtime/ai/`](runtime/ai/) - Add logic to insert `user_feedback` tool call/result messages
4. [`runtime/ai/`](runtime/ai/) - Add attribution prediction prompt and async LLM invocation

### Implementation Flow

```
Frontend submits feedback
    ↓
CompleteStreaming(user_feedback_context: {...})
    ↓
Record user_feedback tool call message
    ↓
If positive: Record simple result → Return
If negative: Return immediately → goroutine runs attribution → stores result
```

## Frontend Implementation

### Components:

1. **Feedback buttons** - Thumbs up/down icons on each AI message bubble
2. **Feedback modal** - Dialog with category checkboxes, comment field, Skip/Submit buttons

### Files to modify:

- `web-common/src/features/chat/` - Add feedback UI components
- Use existing `CompleteStreaming` mutation with new `user_feedback_context` parameter

## Evolution Path

The JSON-based storage provides flexibility to evolve attribution without schema changes:

**v1 (Current):**

- Binary attribution: `rill` vs `project`
- Free-text reasoning + suggested action

**v2 (Future possibilities):**

- Subtypes: `rill.reasoning_error`, `project.missing_measure`, etc.
- Confidence scores: `attribution_confidence: 0.85`
- Related entities: `related_measures: ["churn_rate"]`
- Human override: `human_attribution` field for Rill team corrections

No migrations needed - just add fields to JSON and update prompts.

## Analytics Considerations

### For Use Case 1 (Rill improving AI):

- Daily exports include `ai_messages` table
- Filter by `tool = 'user_feedback'` to extract feedback
- Filter by `predicted_attribution = 'rill'` to focus on AI issues
- Track category distribution over time

### For Use Case 2 (LLM Self-Improvement):

- When building conversation context, `user_feedback` messages are included
- LLM can see patterns like "user marked previous response as incorrect"
- Attribution reasoning provides context for why the failure occurred

### For Use Case 3 (Project admin insights):

- Filter by `predicted_attribution = 'project'` to identify project issues
- Use `suggested_action` to show developers exactly what to fix
- Surface these to project admins via dashboard or notifications
- Track which types of project gaps are most common