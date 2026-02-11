---
name: Short Link Citations Proposal
overview: A design proposal for refactoring citation URLs in Rill's chat system to use short links backed by persisted tool calls instead of embedding query payloads in URLs.
todos: []
---

# Short Link Citations for Rill Chat

## Problem Statement

Citations in AI chat responses are currently embedded as markdown links with URL-encoded query parameters:

```markdown
Revenue grew 25% ([Revenue breakdown](https://ui.rilldata.com/org/proj/-/open-query?query=%7B%22metrics_view%22...%7D))
```

This creates several issues:

- **LLM formatting errors**: The `open_url` contains a full JSON-encoded metrics query, often hundreds of characters of URL-encoded text (`%7B`, `%22`, `%3A`, etc.). The LLM must faithfully reproduce this in its markdown output. In practice, LLMs struggle with nested parentheses in markdown link syntax `([label](url))` and with long opaque strings, causing malformed citations with missing closing parentheses, truncated URLs, or corrupted payloads. (Reported: APP-493)

- **404s from corrupted URLs**: When the LLM mangles the JSON payload in the URL, the `/-/open-query` route receives invalid JSON, causing 404 errors or blank screens. (Reported: APP-493)

- **Hallucinated dimension names**: The LLM sometimes generates `query_metrics_view` calls with dimension names that don't exist, and the citation faithfully links to that broken query. (Reported: APP-493 — this is a tool-call validation problem, orthogonal to citation format, but worth noting.)

- **Frontend complexity**: When the user is already viewing a dashboard, we want citations to update the current view rather than navigate away. The frontend must intercept citation links during markdown rendering, extract the query JSON from the URL, convert it to dashboard URL parameters, and rewrite the link's `href` by hooking into the `marked` library's link renderer — parsing and manipulating HTML strings rather than working with structured data.

- **Embed vs app branching**: In embedded contexts (iframes), standard `<a>` links don't work as expected — cmd/ctrl+click opens a new tab outside the embed, and regular clicks may navigate the parent frame. The current implementation renders citations as `<button>` elements in embedded mode and `<a>` elements in standalone mode, creating two rendering paths in the markdown processor.

- **Limited rendering flexibility**: Design has requested that citations render as icon buttons rather than inline text links. This is difficult to achieve when citations are embedded in markdown strings.

### Root Cause

The fundamental problem is that the query payload travels _through_ the LLM. The backend encodes the full JSON into the URL, hands it to the LLM in a tool result, and hopes the LLM reproduces it faithfully in its markdown output. The LLM is being used as a lossy transport for structured data.

## Proposed Solution: Short Links

Replace the long `open-query?query={json}` URLs with short links that reference persisted tool calls by ID:

```
Before: https://ui.rilldata.com/org/proj/-/open-query?query=%7B%22metrics_view%22%3A%22revenue%22%2C%22dimensions%22%3A%5B%7B%22name%22%3A%22country%22%7D%5D%2C%22measures%22%3A...%7D

After:  https://ui.rilldata.com/org/proj/-/ai/conversations/{conversation_id}/call/{tool_call_id}
```

The query payload bypasses the LLM entirely. The backend stores it (which it already does — every tool call is persisted in the `ai_messages` table), gives the LLM a short reference to it, and resolves the reference deterministically at click time.

### Key Insight

Today, `query_metrics_view` already persists the full query arguments in the `ai_messages` table as part of normal tool call tracking. The tool _call_ message (the assistant's request to invoke `query_metrics_view`) is created and assigned a UUID _before_ the tool handler executes. The short link uses this existing call message ID — the handler receives it, embeds it in the URL, and the query arguments can be retrieved directly from the call message at click time.

### Conversation Forks

`ForkSession` clones messages with **new IDs** (including tool call IDs) and copies content verbatim. Without any additional handling, citation URLs in cloned messages still reference the original conversation and call IDs. This is a known limitation — see Open Questions for options.

## Architecture

```
                    query_metrics_view tool call
                    ┌─────────────────────────────┐
                    │ 0. Call msg persisted w/ UUID│
                    │ 1. Handler receives call ID  │
                    │ 2. Execute query             │
                    │ 3. Return {data, open_url}   │
                    │    where open_url is a       │
                    │    short link with call ID   │
                    └──────────────┬──────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │  LLM emits standard markdown │
                    │  ([label](short_url))         │
                    └──────────────┬──────────────┘
                                   │
              ┌────────────────────┼─────────────────────┐
              ▼                    ▼                      ▼
┌──────────────────┐  ┌────────────────────┐  ┌──────────────────────┐
│ Rill Chat UI     │  │ External MCP       │  │ External MCP         │
│ (today)          │  │ (today: direct     │  │ (future: agent       │
│                  │  │  tool calls)       │  │  exposure)           │
│ Intercepts link, │  │                    │  │                      │
│ resolves tool    │  │ Uses open_url from │  │ LLM emits short link │
│ call from conv   │  │ tool result as-is  │  │ User clicks → browser│
│ in memory.       │  │ (short link works  │  │ → frontend route     │
│ Renders as full  │  │  in any context)   │  │ → gRPC API → resolve │
│ dashboard URL.   │  │                    │  │ → dashboard          │
│ Embed-aware.     │  │                    │  │                      │
└──────────────────┘  └────────────────────┘  └──────────────────────┘
```

## Implementation Approach

### 1. Backend: gRPC Endpoint to Resolve Tool Calls

Add a gRPC endpoint in `proto/rill/runtime/v1/api.proto` alongside the existing AI endpoints (`ListConversations`, `GetConversation`, `Complete`):

```proto
// Resolves an AI tool call to its query arguments.
// Returns the query_metrics_view arguments for the given tool call,
// enabling the frontend to build a dashboard URL.
rpc GetAIToolCall(GetAIToolCallRequest) returns (GetAIToolCallResponse) {
  option (google.api.http) = {
    get: "/v1/instances/{instance_id}/ai/conversations/{conversation_id}/calls/{call_id}"
  };
}

message GetAIToolCallRequest {
  string instance_id = 1;
  string conversation_id = 2;
  string call_id = 3;
}

message GetAIToolCallResponse {
  google.protobuf.Struct query = 1;  // MetricsResolverQuery as JSON
}
```

The handler looks up the tool call message by `call_id`, extracts the `query_metrics_view` arguments from it, and returns them. It verifies that the caller has access to the conversation (either as owner or via sharing) before returning. This prevents unauthorized access to dimension names and filter values.

### 2. Backend: Change `generateOpenURL` to Produce Short Links

In `runtime/ai/metrics_view_query.go`, the `generateOpenURL` function currently serializes the full query as JSON in the URL. Change it to use the conversation ID and tool call message ID:

```go
func (t *QueryMetricsView) generateOpenURL(ctx context.Context, instanceID string, conversationID string, toolCallID string) (string, error) {
    instance, err := t.Runtime.Instance(ctx, instanceID)
    if err != nil {
        return "", fmt.Errorf("failed to get instance: %w", err)
    }
    if instance.FrontendURL == "" {
        return "", nil
    }

    openURL, err := url.Parse(instance.FrontendURL)
    if err != nil {
        return "", fmt.Errorf("failed to parse frontend URL %q: %w", instance.FrontendURL, err)
    }

    openURL.Path, err = url.JoinPath(openURL.Path, "-", "ai", "conversations", conversationID, "call", toolCallID)
    if err != nil {
        return "", fmt.Errorf("failed to join path: %w", err)
    }

    return openURL.String(), nil
}
```

Both IDs are already available when the handler runs. The tool _call_ message (the assistant's request to invoke `query_metrics_view`) is created and assigned a UUID via `uuid.NewString()` in `AddMessage` (`runtime/ai/ai.go`) _before_ the handler executes. The conversation ID is available from the session context. The handler receives the tool call message ID and passes it to `generateOpenURL`.

### 3. Frontend: Add Route for Short Links

Add a SvelteKit route in both `web-local` and `web-admin` at `/-/ai/conversations/{conversation_id}/call/{tool_call_id}`. This route:

1. Calls the `GetAIToolCall` gRPC endpoint (which handles auth/access control)
2. Receives the query arguments
3. Delegates to the existing `openQuery()` function to resolve the explore and redirect to the dashboard

```typescript
// web-admin: src/routes/[organization]/[project]/-/ai/conversations/[conversation_id]/call/[call_id]/+page.ts
export const load: PageLoad = async ({ params, parent }) => {
  await parent();

  // Call the gRPC endpoint to resolve the tool call
  const response = await runtimeServiceGetAIToolCall(
    params.instance_id,
    params.conversation_id,
    params.call_id,
  );

  // Delegate to existing open-query resolution logic
  const query = response.query as MetricsResolverQuery;
  await openQueryFromMetricsResolverQuery({
    query,
    organization: params.organization,
    project: params.project,
  });
};
```

This keeps the `/-/open-query` route unchanged and avoids mixing stateful (call ID lookup) and stateless (inline JSON) operations on the same route. The shared resolution logic (find explore, map query to ExploreState, redirect to dashboard) should be extracted from the existing `openQuery()` function so both routes can use it. The `/-/open-query` route remains for backwards compatibility with any existing bookmarked or shared links that use the old `?query={json}` format.

### 4. Update AI Prompt

No change to the prompt instructions. The `open_url` field in the tool result will now contain a short link instead of a long one, but the LLM's instructions are the same: "use the `open_url` as a markdown link."

```diff
 **Citation Requirements**:
 - Every 'query_metrics_view' result includes an 'open_url' field - use this as a markdown link to cite EVERY quantitative claim made to the user
```

### 5. Frontend: Rill Chat UI — Intercept Short Links

In `AssistantMessage.svelte`, the existing `rewrite-citation-urls.ts` logic changes from "extract JSON from URL and map to dashboard params" to "look up tool call ID from conversation and map to dashboard params."

```typescript
// rewrite-citation-urls.ts (revised)
const CITATION_SHORT_LINK_REGEX =
  /\/-\/ai\/conversations\/([a-f0-9-]+)\/call\/([a-f0-9-]+)\/?$/;

export function getCitationUrlRewriter(
  conversation: Conversation,
  mapper: MetricsResolverQueryToUrlParamsMapper | undefined,
) {
  return (text: string): string | Promise<string> => {
    marked.use({
      renderer: {
        link: (tokens) => {
          const url = URL.parse(tokens.href);
          const match = url?.pathname?.match(CITATION_SHORT_LINK_REGEX);
          if (!match) return false;

          const toolCallID = match[2];
          // Look up the tool call message from the conversation already in memory.
          // This is the assistant's request to invoke query_metrics_view —
          // its content_data contains the query arguments.
          const toolCall = conversation.messages.find(
            (m) =>
              m.id === toolCallID &&
              m.type === "tool_call" &&
              m.tool === "query_metrics_view",
          );
          if (!toolCall) {
            // Fallback: render as standard link (frontend route will handle it)
            return false;
          }

          const queryArgs = JSON.parse(toolCall.content_data);
          const [isValid, urlParams] = mapper(JSON.stringify(queryArgs));
          if (!isValid) return false;

          // Render with full dashboard URL (right-click + copy works)
          return `<a href="?${urlParams.toString()}">${tokens.text}</a>`;
        },
      },
    });
    return marked.parse(text);
  };
}
```

Key behaviors:

- **In Explore context**: Intercepts short links, resolves the tool call from the conversation in memory, renders `<a>` with the full dashboard URL. Right-click + copy gives a self-contained, shareable URL. No network request needed.
- **Outside Explore context**: Short links pass through as standard `<a>` elements pointing to the short URL. Clicking navigates to the frontend route, which calls the gRPC endpoint and redirects.
- **Fallback**: If the tool call ID isn't found in the conversation (edge case), the link renders as-is and the frontend route handles resolution.

Note: the `call_id` here is the tool _call_ message ID — the assistant's request to invoke the tool. Its `content_data` contains the query arguments that were passed to `query_metrics_view`. This is distinct from the tool _result_ message, which contains the query output and is created _after_ the handler returns.

### 6. Frontend: Embed-Aware Rendering (Future)

The existing embed/app branching problem and the icon button rendering request are separate from the short link change. They can be addressed in a follow-up by enhancing the citation link renderer in the `marked` extension — the short link approach doesn't block or change this. With citations now identifiable via a URL pattern (`/-/ai/.../call/...`), the markdown renderer has a clean hook point for custom rendering (buttons in embeds, icon buttons, hover previews, etc.).

## Access Control

The gRPC endpoint verifies that the caller has access to the conversation before returning tool call contents. This prevents unauthorized access to potentially sensitive data (dimension names, filter values, query structure).

- The conversation ID in the URL enables efficient lookup and access control in a single query
- Callers must be the conversation owner or have shared access
- Unauthenticated requests return 401; unauthorized requests return 403
- The Rill Chat UI's in-memory resolution path bypasses this entirely (no network request), since the user already has the conversation loaded

## Compatibility

### External MCP Clients (Today)

External MCP clients (Claude Desktop, Cursor) call `query_metrics_view` directly — they cannot access `router_agent` or `analyst_agent` (gated by `CheckAccess` on user agent prefix in `router_agent.go` and `analyst_agent.go`). They receive `{data, open_url}` in the tool result and use `open_url` as they see fit. The short link works as a standard URL in any context — clicking it opens the frontend route, which resolves the tool call via the gRPC endpoint and redirects to the dashboard.

### External MCP Clients (Future: Agent Exposure)

When agents are exposed to external MCP clients, the agent's markdown output will contain short links. These work as standard URLs — clicking opens the browser, the frontend route calls the gRPC endpoint (with auth), and redirects to the dashboard. No special client-side handling required.

| Consumer                             | How they get citations                                      | Click behavior                              | Right-click copy            |
| ------------------------------------ | ----------------------------------------------------------- | ------------------------------------------- | --------------------------- |
| Rill Chat UI                         | Intercepts short link, resolves from conversation in memory | Native navigation, embed-aware              | Full dashboard URL          |
| External MCP (direct tools)          | `open_url` from tool result                                 | Browser → frontend route → gRPC → dashboard | Short link URL (functional) |
| External MCP (future agent exposure) | Short links in agent markdown output                        | Browser → frontend route → gRPC → dashboard | Short link URL (functional) |

## Alternatives Considered

### 1. XML-style `<cite>` Tags (Previous Version of This Proposal)

Have the LLM output `<cite id="tool_call_id">label</cite>` instead of markdown links. The backend parses cite tags and builds a structured citations array as a sidecar on the message proto.

**Rejected because:**

- Introduces a new output format that only the Rill Chat UI can interpret. External MCP clients (current or future) would see raw `<cite>` tags in the agent's response.
- Requires the backend to regex-parse LLM output and match IDs to tool calls — a fragile contract.
- Requires proto changes (`Citation` message, `citations` field on `Message`).
- The short link approach achieves the same decoupling (query payload doesn't travel through the LLM) while keeping standard markdown link syntax.

### 2. Base64-Encoded Query Payload

Encode the query as proto+base64 instead of JSON+URL-encoding, making the URL shorter and avoiding characters that confuse markdown parsing.

**Rejected because:**

- The payload still travels through the LLM, just in a different encoding. A typical metrics query base64-encodes to 200-500+ opaque characters.
- Base64 corruption is catastrophic — a single wrong character makes the entire payload undecodable. With the current JSON URLs, partial corruption is sometimes recoverable (e.g., stripping a trailing bracket). With base64, there's no graceful degradation.
- Doesn't address the root cause: asking the LLM to be a faithful transport for structured data.

### 3. Extend `/-/open-query` with a `call_id` Parameter

Add `call_id` as an alternative query parameter on the existing `/-/open-query` route, avoiding a new frontend route.

**Not chosen because:**

- `/-/open-query` is stateless today (the query is self-contained in the URL). Adding `call_id` mixes stateful and stateless operations on the same route.
- The two operations are conceptually different: "here's a query, open it" vs. "here's a reference to a query executed in a conversation, resolve it with auth."
- Path-based URLs (`/-/ai/conversations/{conversation_id}/call/{tool_call_id}`) map naturally to the resource hierarchy and keep the conversation ID visible for access control, rather than stuffing both IDs into query parameters.

### 4. Sanitize Malformed URLs (PR #8792)

Parse citation URLs defensively, stripping trailing brackets that the LLM appends (`maybeSanitiseQuery` in PR #8792).

**Assessment:** This is a reasonable near-term stopgap that directly fixes the parenthesis bug with minimal blast radius. It doesn't address the root cause, but could ship ahead of this refactor. The two approaches are complementary, not competing.

## Migration Path

1. **Phase 1: Backend changes**

   - Add `GetAIToolCall` gRPC endpoint in `api.proto` with access control
   - Run `make proto.generate`
   - Implement handler in `runtime/server/`
   - Change `generateOpenURL` to produce short links using conversation ID and tool call message ID
   - Pass the tool call message ID into the handler (it's already available — the call message is created before the handler executes)

2. **Phase 2: Frontend changes**

   - Add `/-/ai/conversations/{conversation_id}/call/{tool_call_id}` route in both `web-local` and `web-admin`
   - Update `rewrite-citation-urls.ts` to recognize short link pattern and resolve tool call from conversation in memory
   - Remove JSON-from-URL extraction logic
   - Existing `/-/open-query` route remains unchanged (no migration needed)

3. **Phase 3: Enhanced rendering (future)**
   - Icon button rendering for citations
   - Embed-aware citation components
   - Hover previews

## Open Questions

1. **Tool call TTL**: How long should tool call messages be retained? If a user shares a citation link months later and the conversation has been garbage collected, the link returns 404. Options: retain indefinitely, set a generous TTL (e.g., 1 year), or accept this as a known limitation with a friendly error page.

2. **Citations in forked conversations**: `ForkSession` clones messages with new IDs but copies content verbatim. Citation URLs in cloned messages still reference the original conversation and call IDs. This works for the forker (who has access to the original), but breaks if the fork is shared with a third party who lacks access to the source conversation.

   Options:

   - **Do nothing (source-bound links)**: Accept that citations in forks reference the original conversation. The forker has access; third-party recipients may not. Simplest, zero implementation cost.
   - **Rewrite at fork time**: During the `ForkSession` clone loop, regex/JSON-rewrite citation URLs in message content to use the new conversation and remapped call IDs (via the existing `oldToNewMessageID` map). Makes forks self-contained, but mutating serialized content strings is fragile.
   - **Stable citation ID store**: Generate a conversation-independent `citation_id` at tool-call time, stored in a dedicated table mapping `{citation_id -> query payload + ACL context}`. Forks copy messages unchanged because the citation ID is stable. Architecturally cleanest, but new storage and access control semantics for a narrow edge case.
   - **Alias table**: On fork, store mapping rows `{fork_conversation_id, old_call_id -> new_call_id}`. The resolver checks the alias table before fetching the tool call. Avoids content mutation but introduces a new lookup path and lifecycle management.
   - **Resolver fallback via lineage**: `GetAIToolCall` follows the `forked_from_session_id` chain to resolve old call IDs from ancestor conversations. No content rewriting or extra tables, but introduces an "implied access" auth pattern that needs careful design.
