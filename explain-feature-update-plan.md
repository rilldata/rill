# Explain Feature Update Plan

This plan covers three areas of improvement to the AI chat ("Explain") feature in the explore interface: enriching chain-of-thought messages, improving context management, and using explore configuration to guide dimension selection.

---

## Current Architecture of the Chat Window

### Where the Code Lives

| Layer | Key Paths | Purpose |
|-------|-----------|---------|
| **Proto/API** | `proto/rill/runtime/v1/api.proto` | RPC definitions for conversation and completion endpoints |
| | `proto/rill/ai/v1/ai.proto` | Core AI message types: `CompletionMessage`, `ContentBlock`, `Tool`, `ToolCall`, `ToolResult` |
| | `proto/rill/admin/v1/ai.proto` | Admin-level low-level completion API |
| **Runtime (Go)** | `runtime/ai/ai.go` | Core session, message, and tool infrastructure (~1600 lines) |
| | `runtime/ai/router_agent.go` | Routes prompts to analyst/developer/feedback agents |
| | `runtime/ai/analyst_agent.go` | Data analysis agent with OODA loop |
| | `runtime/ai/developer_agent.go` | Code/file development agent |
| | `runtime/ai/feedback_agent.go` | Feedback collection agent |
| | `runtime/server/chat.go` | HTTP/gRPC handlers for all chat endpoints (~560 lines) |
| | `runtime/drivers/ai.go` | `AIService` interface definition |
| | `runtime/drivers/openai/openai.go` | OpenAI/Azure provider implementation |
| | `runtime/drivers/claude/claude.go` | Anthropic Claude provider implementation |
| | `runtime/resolvers/ai.go` | YAML resource resolver for AI (report generation) |
| **Frontend** | `web-common/src/features/chat/core/conversation.ts` | Per-conversation state, streaming, cache management |
| | `web-common/src/features/chat/core/conversation-manager.ts` | Multi-conversation lifecycle and selection |
| | `web-common/src/features/chat/core/messages/block-transform.ts` | Transforms `V1Message` → UI blocks |
| | `web-common/src/features/chat/core/input/ChatInput.svelte` | Tiptap-based input with @-mentions |
| | `web-common/src/features/chat/core/messages/Messages.svelte` | Message rendering + auto-scroll |
| | `web-common/src/features/chat/layouts/sidebar/SidebarChat.svelte` | Sidebar layout for explore/canvas dashboards |
| | `web-common/src/features/chat/layouts/fullpage/FullPageChat.svelte` | Full-page layout for dedicated chat |
| | `web-common/src/features/chat/DashboardChat.svelte` | Explore dashboard chat wrapper |
| | `web-common/src/features/chat/ProjectChat.svelte` | Full-page project chat wrapper |
| | `web-common/src/features/dashboards/chat-context.ts` | Builds explore state context for the analyst agent |
| | `web-common/src/runtime-client/sse-fetch-client.ts` | SSE streaming client |

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Frontend (Svelte)                          │
│                                                                     │
│  ┌──────────────┐  ┌──────────────────┐  ┌───────────────────────┐ │
│  │  ChatInput   │  │ ConversationMgr  │  │   Messages / Blocks   │ │
│  │  (Tiptap)    │──│  (TanStack Query │──│  (block-transform.ts) │ │
│  │              │  │   + Svelte stores)│  │                       │ │
│  └──────┬───────┘  └────────┬─────────┘  └───────────────────────┘ │
│         │                   │                                       │
│         ▼                   ▼                                       │
│  ┌─────────────────────────────────────┐                           │
│  │   SSE Fetch Client                  │                           │
│  │   POST /v1/instances/{id}/ai/       │                           │
│  │         complete/stream             │                           │
│  └──────────────┬──────────────────────┘                           │
└─────────────────┼───────────────────────────────────────────────────┘
                  │ SSE (Server-Sent Events)
                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Runtime Server (Go)                               │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  CompleteStreaming Handler (server/chat.go)                  │   │
│  │  - Validates access (UseAI permission)                      │   │
│  │  - Loads or creates Session                                 │   │
│  │  - Subscribes to message stream                             │   │
│  └────────────────────────┬────────────────────────────────────┘   │
│                           │                                         │
│                           ▼                                         │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  AI Session (ai/ai.go)                                      │   │
│  │  - Message storage + dirty tracking                         │   │
│  │  - Tool execution framework                                 │   │
│  │  - Subscriber pattern for streaming                         │   │
│  └────────────────────────┬────────────────────────────────────┘   │
│                           │                                         │
│                           ▼                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐ │
│  │ RouterAgent  │─▶│AnalystAgent  │  │ Tools:                   │ │
│  │              │  │              │  │  query_metrics_view       │ │
│  │              │─▶│DeveloperAgent│  │  get_metrics_view         │ │
│  │              │  │              │  │  list_metrics_views       │ │
│  │              │─▶│FeedbackAgent │  │  navigate, query_sql ...  │ │
│  └──────────────┘  └──────┬───────┘  └──────────────────────────┘ │
│                           │                                         │
│                           ▼                                         │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  AIService Interface (drivers/ai.go)                        │   │
│  │  Complete(ctx, *CompleteOptions) → *CompleteResult           │   │
│  │                                                             │   │
│  │  ┌─────────────────┐        ┌─────────────────┐            │   │
│  │  │ OpenAI Driver    │        │ Claude Driver    │            │   │
│  │  │ (+ Azure OpenAI) │        │ (Anthropic API)  │            │   │
│  │  └─────────────────┘        └─────────────────┘            │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  CatalogStore (persistence)                                 │   │
│  │  - AISession: id, owner, title, timestamps, sharing state   │   │
│  │  - AIMessage: id, parent_id, role, type, content            │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

### Code Path: What Happens When a Prompt Is Submitted

1. **User types a message** → `ChatInput.svelte` captures input, calls `conversation.sendMessage(context, options)`.

2. **Optimistic update** → `Conversation` adds the user message to the TanStack Query cache immediately, sets `isStreaming = true`.

3. **SSE connection opened** → `SSEFetchClient` sends `POST /v1/instances/{instanceId}/ai/complete/stream` with JSON body containing `prompt`, `conversationId`, `agent`, and context fields (metrics view, filters, time range, etc.).

4. **Server handler** (`CompleteStreamingHandler` in `server/chat.go`) → validates `UseAI` permission, parses request, loads or creates an `ai.Session`.

5. **Session subscription** → server calls `session.Subscribe()` to get a channel that receives messages as they're produced, then launches the agent call in a background goroutine.

6. **Router agent** (`router_agent.go`) → receives the prompt, decides which specialized agent to invoke (analyst, developer, or feedback) based on LLM classification.

7. **Agent execution** (e.g., `analyst_agent.go`) →
   - Pre-invokes context tools (`get_metrics_view`, etc.) to gather data definitions.
   - Builds a system prompt with project instructions, available tools, and dashboard context.
   - Enters the **LLM completion loop** (`session.Complete()`):
     - Converts session messages to protobuf `CompletionMessage` format.
     - Calls `AIService.Complete()` on the configured provider (OpenAI/Claude).
     - If the LLM returns tool calls: executes each tool, records call/result messages, loops.
     - If the LLM returns text: records the final response, exits loop.

8. **Message streaming** → each message added to the session (tool calls, progress, results, final text) is pushed to subscribers. The server goroutine reads from the subscription channel and writes SSE events to the HTTP response.

9. **Frontend processes SSE events** → `Conversation` receives each `V1CompleteStreamingResponse`, skips echoed user messages, updates the TanStack Query cache, triggers tool callbacks (e.g., navigate to a chart).

10. **Block transformation** → `block-transform.ts` converts the flat message list into typed UI blocks: `TextBlock`, `ThinkingBlock`, `ChartBlock`, `FileDiffBlock`, `WorkingBlock`.

11. **Rendering** → `Messages.svelte` renders blocks via `AssistantMessage`, `ThinkingBlock`, `ChartBlock`, etc. Auto-scroll keeps the latest content visible.

12. **Flush and cleanup** → after the agent completes, `session.Flush()` persists all dirty messages to the `CatalogStore`. The SSE connection closes. Frontend sets `isStreaming = false`.

### How Conversations Are Stored

**Backend (CatalogStore):**
- **`AISession`** table: `id`, `instance_id`, `owner_id`, `title`, `user_agent`, `shared_until_message_id`, `forked_from_session_id`, `created_on`, `updated_on`.
- **`AIMessage`** table: `id`, `parent_id`, `session_id`, `time`, `index` (ordering), `role`, `type`, `tool`, `content_type`, `content`.
- Messages are kept in memory during execution (marked `dirty = true`) and batch-flushed to the catalog after the agent completes.
- Sessions support sharing (up to a specific message) and forking (copying messages to a new session).

**Frontend (dual persistence):**
- **Full-page chat** (`URLConversationSelector`): conversation ID stored in the URL path (`/-/ai/{conversationId}`), enabling shareable links and browser history navigation. Last-visited conversation persisted in `sessionStorage`.
- **Sidebar chat** (`BrowserStorageConversationSelector`): conversation ID stored in `sessionStorage` keyed by `sidebar-conversation-id-{org}-{project}`. Cleared when switching projects.
- **Message data**: cached in TanStack Query with the conversation ID as the query key. Updated optimistically on send and incrementally during streaming.

### Important Types and Data Structures

**Backend (Go):**

| Type | Location | Purpose |
|------|----------|---------|
| `ai.Runner` | `runtime/ai/ai.go` | Top-level manager; registers tools, creates sessions |
| `ai.BaseSession` | `runtime/ai/ai.go` | Core session: messages slice, subscribers map, dirty tracking |
| `ai.Session` | `runtime/ai/ai.go` | Extends BaseSession with `ParentID` for nested tool call hierarchy |
| `ai.Message` | `runtime/ai/ai.go` | In-memory message: `ID`, `ParentID`, `Role`, `Type`, `Tool`, `ContentType`, `Content` |
| `ai.CompiledTool` | `runtime/ai/ai.go` | Tool definition: `Name`, `Spec`, `CheckAccess`, `JSONHandler` |
| `drivers.AIService` | `runtime/drivers/ai.go` | Provider interface: single `Complete()` method |
| `drivers.CompleteOptions` | `runtime/drivers/ai.go` | LLM call input: `Messages`, `Tools`, `OutputSchema` |
| `drivers.CompleteResult` | `runtime/drivers/ai.go` | LLM call output: `Message`, `InputTokens`, `OutputTokens` |
| `drivers.AISession` | `runtime/drivers/catalog.go` | Database DTO for session metadata |
| `drivers.AIMessage` | `runtime/drivers/catalog.go` | Database DTO for persisted messages |
| `aiv1.CompletionMessage` | `proto/rill/ai/v1/ai.proto` | Protobuf message format: `role` + repeated `ContentBlock` |
| `aiv1.ContentBlock` | `proto/rill/ai/v1/ai.proto` | Oneof: `text` \| `tool_call` \| `tool_result` |

**Frontend (TypeScript):**

| Type | Location | Purpose |
|------|----------|---------|
| `Conversation` | `conversation.ts` | Per-conversation state, streaming, cache, forking |
| `ConversationManager` | `conversation-manager.ts` | Multi-conversation lifecycle, max 3 concurrent streams |
| `Block` (union) | `block-transform.ts` | `TextBlock` \| `ThinkingBlock` \| `ChartBlock` \| `FileDiffBlock` \| `WorkingBlock` \| `SimpleToolCall` |
| `SSEFetchClient` | `sse-fetch-client.ts` | SSE protocol client with `start()`, `stop()`, `on()` |
| `ChatConfig` | `types.ts` | Configuration: agent name, context getter, labels, placeholders |
| `InlineContextType` | `inline-context.ts` | Enum: `MetricsView`, `Canvas`, `Dimension`, `Measure`, etc. |

**Message Roles and Types:**
- Roles: `system`, `user`, `assistant`, `tool`
- Types: `call` (tool invocation), `result` (tool output), `progress` (streaming updates)
- Content types: `text`, `json`, `error`

### LLM Provider Abstraction

The system uses a clean driver-based abstraction to support multiple LLM providers:

**Interface** (`runtime/drivers/ai.go`):
```go
type AIService interface {
    Complete(ctx context.Context, opts *CompleteOptions) (*CompleteResult, error)
}
```

All communication with LLMs flows through this single method. Messages use Rill's protobuf format (`aiv1.CompletionMessage` with `ContentBlock` unions); each driver handles conversion to/from the provider's native format.

**Implementations:**

| Driver | File | Default Model | Key Details |
|--------|------|---------------|-------------|
| `openai` | `runtime/drivers/openai/openai.go` | `gpt-5-2` | Supports Azure OpenAI via `api_type`; handles asymmetric tool call format (multiple calls in one assistant message, separate tool result messages) |
| `claude` | `runtime/drivers/claude/claude.go` | `claude-opus-4-5-20251101` | Uses Anthropic Beta Messages API; system messages separated from conversation; tool results sent as user messages per Anthropic API convention |
| `mock/ai` | `runtime/drivers/mock/ai/ai.go` | N/A | Testing only; echoes or returns mock tool calls |

**Driver selection:** configured per instance via `ai_connector` in `rill.yaml` (project-level) or via instance configuration (environment-level). The runtime resolves the connector name → acquires a driver handle → calls `handle.AsAI()` to get the `AIService`.

**Provider-specific details:**
- **OpenAI**: configurable `model`, `max_output_tokens`, `reasoning_effort`, `base_url`, `api_key`. Temperature defaults to 0.1 (omitted when reasoning is enabled).
- **Claude**: configurable `model`, `max_tokens`, `temperature`, `base_url`, `api_key`. Supports structured outputs via `structured-outputs-2025-11-13` beta.
- **Azure OpenAI**: configured by setting `api_type: "azure"` on the OpenAI driver with appropriate `base_url` and optional `api_version`.

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
