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

The fundamental problem is that the query payload travels *through* the LLM. The backend encodes the full JSON into the URL, hands it to the LLM in a tool result, and hopes the LLM reproduces it faithfully in its markdown output. The LLM is being used as a lossy transport for structured data.

## Proposed Solution: Short Links

Replace the long `open-query?query={json}` URLs with short links that reference persisted tool calls by ID:

```
Before: https://ui.rilldata.com/org/proj/-/open-query?query=%7B%22metrics_view%22%3A%22revenue%22%2C%22dimensions%22%3A%5B%7B%22name%22%3A%22country%22%7D%5D%2C%22measures%22%3A...%7D

After:  https://ui.rilldata.com/org/proj/-/ai/call/550e8400-e29b-41d4-a716-446655440000
```

The query payload bypasses the LLM entirely. The backend stores it (which it already does — every tool call is persisted in the `ai_messages` table), gives the LLM a short reference to it, and resolves the reference deterministically at click time.

### Key Insight

Today, `query_metrics_view` already persists the full query arguments in the `ai_messages` table as part of normal tool call tracking. The short link approach simply uses this existing data as the source of truth for citations, rather than re-serializing the query into a URL and asking the LLM to relay it.

## Architecture

```
                    query_metrics_view tool call
                    ┌─────────────────────────────┐
                    │ 1. Execute query             │
                    │ 2. Persist call + args       │
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
│ in memory.       │  │ (short link works  │  │ → backend redirect   │
│ Renders as full  │  │  in any context)   │  │ → dashboard          │
│ dashboard URL.   │  │                    │  │                      │
│ Embed-aware.     │  │                    │  │                      │
└──────────────────┘  └────────────────────┘  └──────────────────────┘
```

## Implementation Approach

### 1. Backend: Change `generateOpenURL` to Produce Short Links

In `runtime/ai/metrics_view_query.go`, the `generateOpenURL` function currently serializes the full query as JSON in the URL. Change it to use the tool call's message ID:

```go
func (t *QueryMetricsView) generateOpenURL(ctx context.Context, instanceID string, messageID string) (string, error) {
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

    openURL.Path, err = url.JoinPath(openURL.Path, "-", "ai", "call", messageID)
    if err != nil {
        return "", fmt.Errorf("failed to join path: %w", err)
    }

    return openURL.String(), nil
}
```

The message ID is already available — each tool call message gets a UUID when created (`uuid.NewString()` in `runtime/ai/ai.go`). The function signature changes to accept the message ID instead of the query map.

### 2. Backend: Add Redirect Endpoint

Add an HTTP endpoint that resolves a tool call ID to the existing `/-/open-query` route. This keeps all state management (finding the explore, mapping query to dashboard state) in the existing `open-query` infrastructure.

```
GET /-/ai/call/{tool_call_id}
  → Look up tool call message by ID
  → Extract query args from message content
  → Build /-/open-query?query={json} URL
  → 302 redirect
```

```go
// In runtime/server/ — register alongside existing AI routes
func (s *Server) handleAICallRedirect(w http.ResponseWriter, r *http.Request) {
    toolCallID := r.PathValue("tool_call_id")

    // Look up the tool call message
    msg, err := s.findAIMessage(r.Context(), toolCallID)
    if err != nil || msg.Tool != "query_metrics_view" {
        http.NotFound(w, r)
        return
    }

    // Build the open-query redirect URL
    redirectURL := buildOpenQueryURL(r, msg.Content)
    http.Redirect(w, r, redirectURL, http.StatusFound)
}
```

This endpoint is intentionally thin — it's a lookup and redirect. The `open-query` route handles all the complex state resolution (finding the explore for the metrics view, converting the query to ExploreState, generating the final dashboard URL).

**Note:** This requires adding a `FindAIMessage(ctx, messageID)` method to the catalog interface. Currently, messages can only be queried by session ID. A direct lookup by message ID is needed for this endpoint.

### 3. Update AI Prompt

The prompt change is minimal. The citation format stays as standard markdown links — only the URL changes:

```diff
 **Citation Requirements**:
 - Every 'query_metrics_view' result includes an 'open_url' field - use this as a markdown link to cite EVERY quantitative claim made to the user
```

No change to the prompt instructions. The `open_url` field in the tool result will now contain a short link instead of a long one, but the LLM's instructions are the same: "use the `open_url` as a markdown link."

### 4. Frontend: Rill Chat UI — Intercept Short Links

In `AssistantMessage.svelte`, the existing `rewrite-citation-urls.ts` logic changes from "extract JSON from URL and map to dashboard params" to "look up tool call ID from conversation and map to dashboard params."

```typescript
// rewrite-citation-urls.ts (revised)
const CITATION_SHORT_LINK_REGEX = /\/-\/ai\/call\/([a-f0-9-]+)\/?$/;

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

          const toolCallID = match[1];
          // Look up the tool call from the conversation already in memory
          const toolCall = conversation.messages.find(
            (m) => m.id === toolCallID && m.tool === "query_metrics_view"
          );
          if (!toolCall) {
            // Fallback: render as standard link (backend redirect will handle it)
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
- **In Explore context**: Intercepts short links, resolves the tool call from the conversation in memory, renders `<a>` with the full dashboard URL. Right-click + copy gives a self-contained, shareable URL.
- **Outside Explore context**: Short links pass through as standard `<a>` elements pointing to the short URL. Clicking triggers the backend redirect.
- **Fallback**: If the tool call ID isn't found in the conversation (edge case), the link renders as-is. Clicking follows the backend redirect endpoint.

### 5. Frontend: Embed-Aware Rendering (Future)

The existing embed/app branching problem and the icon button rendering request are separate from the short link change. They can be addressed in a follow-up by enhancing the citation link renderer in the `marked` extension — the short link approach doesn't block or change this. With citations now identifiable via a URL pattern (`/-/ai/call/`), the markdown renderer has a clean hook point for custom rendering (buttons in embeds, icon buttons, hover previews, etc.).

## Compatibility

### External MCP Clients (Today)

External MCP clients (Claude Desktop, Cursor) call `query_metrics_view` directly — they cannot access `router_agent` or `analyst_agent` (gated by `CheckAccess` on user agent prefix in `router_agent.go` and `analyst_agent.go`). They receive `{data, open_url}` in the tool result and use `open_url` as they see fit. The short link works as a standard URL in any context.

### External MCP Clients (Future: Agent Exposure)

When agents are exposed to external MCP clients, the agent's markdown output will contain short links. These work as standard URLs — clicking opens the browser, the backend resolves the tool call, and redirects to the dashboard. No special client-side handling required.

| Consumer | How they get citations | Click behavior | Right-click copy |
|---|---|---|---|
| Rill Chat UI | Intercepts short link, resolves from conversation in memory | Native navigation, embed-aware | Full dashboard URL |
| External MCP (direct tools) | `open_url` from tool result | Browser → backend redirect → dashboard | Short link URL (functional) |
| External MCP (future agent exposure) | Short links in agent markdown output | Browser → backend redirect → dashboard | Short link URL (functional) |

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

### 3. Sanitize Malformed URLs (PR #8792)

Parse citation URLs defensively, stripping trailing brackets that the LLM appends (`maybeSanitiseQuery` in PR #8792).

**Assessment:** This is a reasonable near-term stopgap that directly fixes the parenthesis bug with minimal blast radius. It doesn't address the root cause, but could ship ahead of this refactor. The two approaches are complementary, not competing.

## Migration Path

1. **Phase 1: Backend changes**
   - Add `FindAIMessage(ctx, messageID)` method to the catalog interface
   - Add `/-/ai/call/{tool_call_id}` redirect endpoint
   - Change `generateOpenURL` to produce short links using the tool call message ID
   - Add frontend route for `/-/ai/call/{tool_call_id}` (delegates to backend redirect for direct navigation, or is intercepted by Rill Chat UI)

2. **Phase 2: Frontend changes**
   - Update `rewrite-citation-urls.ts` to recognize short link pattern and resolve from conversation in memory
   - Remove JSON-from-URL extraction logic
   - Existing `open-query` route remains as the backend redirect target (no changes needed)

3. **Phase 3: Enhanced rendering (future)**
   - Icon button rendering for citations
   - Embed-aware citation components
   - Hover previews

## Open Questions

1. **Tool call TTL**: How long should tool call messages be retained? If a user shares a citation link months later and the session has been garbage collected, the link returns 404. Options: retain indefinitely, set a generous TTL (e.g., 1 year), or accept this as a known limitation with a friendly error page.
