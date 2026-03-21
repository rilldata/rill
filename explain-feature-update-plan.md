# Explain Feature Update Plan

This plan covers three areas of improvement to the AI chat ("Explain") feature in the explore interface: enriching chain-of-thought messages, improving context management, and using explore configuration to guide dimension selection.

---

## 1. Enriching Chain-of-Thought Messages

### Current State

- The analyst agent executes tool calls (e.g., `query_metrics_view`, `get_metrics_view`) and the frontend groups them into collapsible "thinking blocks" (`ThinkingBlock` in `web-common/src/features/chat/core/messages/thinking/thinking-block.ts`).
- Tool display names come from backend `meta` fields (e.g., `"openai/toolInvocation/invoking": "Analyzing..."`) — these are static, generic labels.
- The LLM produces `<thinking>` tag content between queries, but this is embedded in the response text; there is no structured "reasoning step" message type surfaced in the thinking block UI.
- The thinking block only shows a duration and a list of tool calls with expand/collapse. Users cannot see **why** the agent chose a particular query or what it learned from results.

### Proposed Changes

#### 1a. Add structured progress messages with reasoning context

**Backend** (`runtime/ai/analyst_agent.go`, `runtime/ai/ai.go`):
- After each tool call result in the analyst agent's OODA loop, emit a `progress` message summarizing what was learned and what the next step is. The LLM already generates `<thinking>` blocks; parse these and emit them as structured progress messages with a new `content_type` value (e.g., `"reasoning"`) so the frontend can distinguish reasoning from plain text progress.
- Add a field to the `Message` struct (or use `content_data` JSON) to carry a `step_label` — a short human-readable label like "Examining revenue trends by country" or "Found anomaly in Q4 data, drilling deeper."

**Frontend** (`web-common/src/features/chat/core/messages/thinking/`):
- Create a new `ReasoningStep.svelte` component that renders reasoning progress messages within thinking blocks with:
  - A brief label (the `step_label`)
  - Expandable detail showing the full reasoning text
  - Visual indicator of the step's position in the analysis sequence (e.g., step 1 of N)
- Update `block-transform.ts` to route `"reasoning"` content type messages into thinking blocks with the new rendering.

#### 1b. Enrich tool call display names with dynamic context

**Backend** (`runtime/ai/metrics_view_query.go`, `runtime/ai/metrics_view_summary.go`):
- Include contextual information in tool call progress messages. For example, when `query_metrics_view` is called, emit a progress message like "Querying revenue by country for Q4 2025" instead of the generic "Analyzing...".
- Update each tool's `Meta` to include a template pattern (e.g., `"Querying {{measures}} by {{dimensions}}"`) that the frontend can interpolate from `contentData`.

**Frontend** (`web-common/src/features/chat/core/messages/tools/tool-display-names.ts`):
- Extend `getToolDisplayName` to accept the tool call's `contentData` and interpolate dynamic values (dimension names, measure names, time ranges) into the display string.
- Fall back to the static meta name when `contentData` is unavailable.

#### 1c. Show analysis phase indicators

**Frontend** (`web-common/src/features/chat/core/messages/thinking/`):
- Add a phase indicator to thinking blocks mapping to the analyst agent's three phases:
  - **Phase 1: Discovery** — when the agent is fetching metrics view definitions
  - **Phase 2: Analysis** — when the agent is running queries in the OODA loop
  - **Phase 3: Visualization** — when the agent is creating charts
- Determine the phase from the tool names in the thinking block (e.g., `list_metrics_views`/`get_metrics_view` → Discovery, `query_metrics_view` → Analysis, `create_chart` → Visualization).

### Key Files

| File | Change |
|------|--------|
| `runtime/ai/analyst_agent.go` | Emit reasoning progress messages, add step labels |
| `runtime/ai/ai.go` | Support `"reasoning"` content type in progress messages |
| `web-common/src/features/chat/core/messages/thinking/thinking-block.ts` | Add reasoning step data to ThinkingBlock type |
| `web-common/src/features/chat/core/messages/thinking/ReasoningStep.svelte` | **New**: reasoning step component |
| `web-common/src/features/chat/core/messages/block-transform.ts` | Route reasoning messages into thinking blocks |
| `web-common/src/features/chat/core/messages/tools/tool-display-names.ts` | Dynamic display name interpolation |
| `web-common/src/features/chat/core/types.ts` | Add `"reasoning"` content type constant |

---

## 2. Improving Context Management in Chat

### Current State

- `Conversation` class (`web-common/src/features/chat/core/conversation.ts`) manages per-conversation state, sends messages via SSE, and caches responses with TanStack Query.
- The `ChatConfig.additionalContextStoreGetter` pattern passes explore state (filters, time range, visible dimensions/measures) as a `Partial<RuntimeServiceCompleteBody>` on every message send.
- Context is captured at message-send time as a snapshot. If the user changes dashboard state between messages, the next message picks up the new state, but the conversation history retains stale context from prior messages.
- The backend `AnalystAgent.Handler` rebuilds the system prompt on every invocation but re-uses messages from the session history, which may reference dimensions/measures that are no longer visible.
- There is no mechanism to inform the agent that the user's dashboard context changed between messages.

### Proposed Changes

#### 2a. Track context changes across messages

**Frontend** (`web-common/src/features/dashboards/chat-context.ts`):
- Maintain a reference to the previously sent context. When the user sends a new message, diff the current explore state against the previous state.
- If the context changed (different filters, time range, visible dimensions), include a structured `context_delta` in the request body or prepend a system note to the prompt (e.g., "Note: The user has changed the time range from X to Y and added a filter on country='US'").

**Backend** (`runtime/ai/analyst_agent.go`):
- When the agent receives a context delta or detects that dimensions/measures/filters changed from a prior invocation, inject a system message into the conversation acknowledging the change: "The user has updated their dashboard view. The following changes were made: [...]"
- This helps the LLM avoid referencing stale data from prior tool call results.

#### 2b. Prune or summarize stale tool results

**Backend** (`runtime/ai/analyst_agent.go`):
- When building `messages` for the LLM completion call, identify tool results from prior invocations that used different filters or time ranges than the current context.
- Instead of including full stale results, replace them with a summary: "Previous query result (different time range: X–Y): [summary]".
- This reduces context window usage and prevents the LLM from citing outdated data.

#### 2c. Improve conversation forking context preservation

**Backend** (`runtime/server/chat.go`, `runtime/ai/ai.go`):
- When a conversation is forked (`ForkConversation`), carry forward a context summary of the original conversation's state (explore name, key filters, time range) so the forked conversation starts with proper context even if the new user's dashboard state differs.

### Key Files

| File | Change |
|------|--------|
| `web-common/src/features/dashboards/chat-context.ts` | Track previous context, compute deltas |
| `web-common/src/features/chat/core/conversation.ts` | Pass context delta in message sends |
| `runtime/ai/analyst_agent.go` | Inject context change notifications, prune stale results |
| `runtime/ai/ai.go` | Support context delta in session metadata |
| `proto/rill/runtime/v1/api.proto` | Add optional `context_delta` field to `CompleteStreamingRequest` |
| `runtime/server/chat.go` | Handle context delta in Complete handlers, improve fork context |

---

## 3. Explore Configuration-Guided Dimension Selection

### Current State

- The explore dashboard defines a `default_preset` (`ExplorePreset` in `resources.proto`) which specifies default dimensions and measures to show.
- The frontend already sends `visibleDimensions` and `visibleMeasures` from the dashboard state to the API via `V1AnalystAgentContext` — but only when not all dimensions are visible (`!allDimensionsVisible`).
- The analyst agent receives these as `args.Dimensions` and includes them in the system prompt as a hint, but the agent is free to ignore them.
- The agent currently has no tool to discover what dimensions are available vs. visible vs. in the default preset — it relies on `get_metrics_view` which returns all dimensions regardless of explore configuration.

### Proposed Changes

#### 3a. Pass explore preset and visibility context to the agent

**Frontend** (`web-common/src/features/dashboards/chat-context.ts`):
- Always send the full list of `visibleDimensions` and `visibleMeasures` to the API, even when all are visible, along with a flag `allDimensionsVisible` and `allMeasuresVisible`.
- Also send the explore's `default_preset` dimensions (from the explore spec) so the agent knows what the project creator intended as the default view.

**Proto** (`proto/rill/runtime/v1/api.proto`):
- Add fields to `AnalystAgentContext`:
  ```protobuf
  // Whether all dimensions are currently visible in the dashboard.
  bool all_dimensions_visible = 14;
  // Whether all measures are currently visible in the dashboard.
  bool all_measures_visible = 15;
  // Default dimensions from the explore's default_preset.
  repeated string default_dimensions = 16;
  // Default measures from the explore's default_preset.
  repeated string default_measures = 17;
  ```

#### 3b. Update the analyst agent's system prompt with dimension guidance

**Backend** (`runtime/ai/analyst_agent.go`):
- Update `systemPrompt()` to include clear guidance on dimension selection:
  - If the user has filtered to specific dimensions (`dimensions` is non-empty and `all_dimensions_visible` is false): "The user is currently viewing these dimensions: [...]. Prioritize these in your analysis. You may query other dimensions if directly relevant to the user's question."
  - If all dimensions are visible and default dimensions exist: "The project's default dimensions are: [...]. Start your analysis with these and expand to other dimensions as needed."
  - If all dimensions are visible and no default preset: "All dimensions are available. Use `get_metrics_view` to see the full list and select the most relevant ones for the analysis."

#### 3c. Add a `list_dimensions` tool for explicit dimension discovery

**Backend** (`runtime/ai/list_dimensions.go` — **new file**):
- Create a new tool `list_dimensions` that returns dimensions categorized by their status:
  ```json
  {
    "all_dimensions": ["country", "city", "product", "channel", ...],
    "default_dimensions": ["country", "product"],
    "visible_dimensions": ["country", "city"],
    "recommended": "Use visible_dimensions for the current analysis. Expand to default_dimensions or all_dimensions if the user's question requires broader exploration."
  }
  ```
- This tool reads from the explore spec's `default_preset` and the current agent args to provide a curated list.
- Register the tool in `NewRunner()` and make it available to the analyst agent.

**Backend** (`runtime/ai/analyst_agent.go`):
- Add `list_dimensions` to the analyst agent's available tools.
- Update the system prompt to instruct the agent: "Before querying, consider calling `list_dimensions` to understand which dimensions are most relevant to the current dashboard view."

#### 3d. Dimension selection decision tree in the system prompt

Add clear instructions to the analyst agent's system prompt for how to choose dimensions:

```
<dimension_selection>
When selecting dimensions for your analysis:
1. If the user explicitly mentions dimensions in their question, use those.
2. If the dashboard has specific visible dimensions set by the user, prioritize those.
3. If the dashboard has a default preset with specific dimensions, start with those.
4. Otherwise, call `list_dimensions` to discover available dimensions and select
   the most analytically relevant ones based on the user's question.

Avoid querying all dimensions at once — be selective and purposeful.
</dimension_selection>
```

### Key Files

| File | Change |
|------|--------|
| `proto/rill/runtime/v1/api.proto` | Add visibility/default fields to `AnalystAgentContext` |
| `web-common/src/features/dashboards/chat-context.ts` | Send visibility flags and default preset dimensions |
| `runtime/ai/analyst_agent.go` | Update system prompt with dimension guidance, add `list_dimensions` tool |
| `runtime/ai/list_dimensions.go` | **New**: `list_dimensions` tool implementation |
| `runtime/ai/ai.go` | Register `list_dimensions` tool |

---

## Implementation Order

1. **Phase 1 — Dimension selection (Section 3)**: Start here as it changes the proto and provides immediate value by improving analysis relevance.
   - 3a: Proto + frontend context changes
   - 3b: System prompt updates
   - 3c: `list_dimensions` tool
   - 3d: Decision tree in system prompt

2. **Phase 2 — Context management (Section 2)**: Build on the proto changes from Phase 1.
   - 2a: Context delta tracking
   - 2b: Stale result pruning
   - 2c: Fork context preservation

3. **Phase 3 — Chain-of-thought enrichment (Section 1)**: This is the most UI-heavy work and benefits from the backend changes in Phases 1–2.
   - 1a: Reasoning progress messages
   - 1b: Dynamic tool display names
   - 1c: Phase indicators

---

## Testing Strategy

- **Unit tests**: Add tests for `list_dimensions` tool, context delta computation, and dynamic display name interpolation.
- **Integration tests**: Extend `analyst_agent_test.go` to verify dimension selection behavior with different explore presets.
- **Frontend unit tests**: Test `block-transform.ts` changes with reasoning message types, test `chat-context.ts` delta computation.
- **Manual testing**: Verify the enriched thinking blocks render correctly in both `web-local` and `web-admin` explore dashboards, including embedded contexts.
