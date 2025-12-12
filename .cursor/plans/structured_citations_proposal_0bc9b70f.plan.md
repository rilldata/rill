---
name: Structured Citations Proposal
overview: A design proposal for refactoring citation handling in Rill's chat system to use structured data instead of embedded markdown URLs.
todos: []
---

# Structured Citations for Rill Chat

## Problem Statement

Citations in AI chat responses are currently embedded as markdown links with URL-encoded query parameters:

```markdown
Revenue grew 25% ([Revenue breakdown](https://ui.rilldata.com/org/proj/-/open-query?query=%7B%22metrics_view%22...%7D))
```

This creates several issues:

- **Frontend complexity**: The AI embeds citation URLs as standard markdown links pointing to a dedicated frontend route (`/-/open-query?query=...`). However, when the user is already viewing a dashboard, we want citations to update the current view rather than navigate away. This means the frontend must intercept these links during markdown rendering, extract the query JSON from the URL, convert it to the current dashboard's URL parameters, and rewrite the link's `href`. This is done by hooking into the `marked` library's link renderer, which means we're parsing and manipulating HTML strings rather than working with structured data.

- **Embed vs app branching**: In embedded contexts (iframes), standard `<a>` links don't work as expected - cmd/ctrl+click opens a new tab outside the embed, and regular clicks may navigate the parent frame. To work around this, the current implementation renders citations as `<button>` elements in embedded mode and `<a>` elements in standalone mode. This creates two rendering paths in the markdown processor, requires event delegation to handle button clicks, and results in buttons that lack standard link affordances (no URL preview on hover, no right-click context menu).

- **LLM formatting errors**: LLMs struggle with nested parentheses in markdown link syntax `([label](url))`, causing malformed citations with missing closing parentheses.

- **Limited rendering flexibility**: Design has requested that citations render as icon buttons rather than inline text links. This is difficult to achieve when citations are embedded in markdown strings - we'd need to post-process HTML to replace link elements with icon markup, adding yet another transformation step.

## Proposed Solution: Structured Content Model

Have the backend return citations as structured data, separate from the markdown content:

```typescript
interface Message {
  content: string;  // "Revenue grew 25% <cite id="abc123">see breakdown</cite>..."
  citations: Array<{
    id: string;      // Tool call ID
    text: string;    // Label from the cite tag
    query: MetricsResolverQuery;  // Raw query object
  }>;
}
```

The AI uses a simple XML-style syntax that's consistent with existing prompt patterns:

```
Revenue increased by 25% <cite id="abc123">see regional breakdown</cite> 
compared to the previous quarter.
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        MCP Layer (Universal)                     │
│  ┌─────────────────────┐    ┌─────────────────────────────────┐ │
│  │ query_metrics_view  │───▶│ Result: {data, open_url}        │ │
│  └─────────────────────┘    └─────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
           │                              │
           ▼                              ▼
┌──────────────────────┐      ┌──────────────────────────────────┐
│  External MCP Clients │      │         Rill Chat UI             │
│  (Claude Desktop,     │      │                                  │
│   Cursor, etc.)       │      │  Complete API                    │
│                       │      │       │                          │
│  Uses open_url in     │      │       ▼                          │
│  markdown links       │      │  Collect tool calls              │
│       │               │      │       │                          │
│       ▼               │      │       ▼                          │
│  Links to Rill UI     │      │  Build citations array           │
│  (external browser)   │      │       │                          │
│                       │      │       ▼                          │
│                       │      │  CitationLink component          │
│                       │      │  (native navigation, embed-aware)│
└──────────────────────┘      └──────────────────────────────────┘
```

## Implementation Approach

### 1. Proto Changes

Add citations as a first-class field on the `Message` proto:

```proto
message Citation {
  string id = 1;     // Tool call ID
  string text = 2;   // Label text
  google.protobuf.Struct query = 3;  // MetricsResolverQuery as JSON
}

message Message {
  // ... existing fields ...
  repeated Citation citations = 12;
}
```

### 2. Backend: Collect Citations from Tool Calls

The session already tracks all tool calls. After completion:

1. Parse `<cite id="...">...</cite>` tags from the response text
2. Match each `id` to the corresponding `query_metrics_view` tool call
3. Build the citations array with the query args from each matched tool call
```go
func buildCitationsFromResponse(s *ai.Session, responseText string) []Citation {
    // Parse <cite id="...">label</cite> tags
    citeRegex := regexp.MustCompile(`<cite id="([^"]+)">([^<]+)</cite>`)
    matches := citeRegex.FindAllStringSubmatch(responseText, -1)
    
    var citations []Citation
    for _, match := range matches {
        toolCallID := match[1]
        label := match[2]
        
        // Find the tool call message with this ID
        callMsg, ok := s.Message(ai.FilterByID(toolCallID))
        if !ok || callMsg.Tool != ai.QueryMetricsViewName {
            continue
        }
        
        var args ai.QueryMetricsViewArgs
        json.Unmarshal([]byte(callMsg.Content), &args)
        
        citations = append(citations, Citation{
            ID:    toolCallID,
            Text:  label,
            Query: args,
        })
    }
    return citations
}
```


### 3. Update AI Prompt

Change analyst agent instructions to use the XML-style cite syntax:

```diff
- Every 'query_metrics_view' result includes an 'open_url' field - use this as a markdown link
+ Every 'query_metrics_view' result includes a tool call ID. Reference queries using: <cite id="TOOL_CALL_ID">descriptive label</cite>
```

This syntax is consistent with existing prompt patterns like `<role>`, `<process>`, `<example>`.

### 4. Frontend: Rendering Citations within Markdown

There's a gap between markdown rendering (which outputs HTML strings) and Svelte components. We'll use a two-phase approach:

#### Phase 1: Action-Enhanced HTML (simpler, immediate)

Keep using `marked` but extend it to recognize cite tags and output HTML with data attributes. A Svelte action then "hydrates" these elements with navigation behavior.

```typescript
// marked extension - outputs HTML placeholders
marked.use({
  extensions: [{
    name: 'cite',
    level: 'inline',
    start(src) { return src.match(/<cite/)?.index; },
    tokenizer(src) {
      const match = src.match(/^<cite id="([^"]+)">([^<]+)<\/cite>/);
      if (match) {
        return { type: 'cite', raw: match[0], id: match[1], text: match[2] };
      }
    },
    renderer(token) {
      return `<a href="#" data-citation-id="${token.id}" class="citation-link">${token.text}</a>`;
    }
  }]
});
```
```svelte
<!-- AssistantMessage.svelte -->
<script>
  import { enhanceCitations } from "./enhance-citations-action";
  export let content: string;
  export let citations: Citation[];
</script>

<div use:enhanceCitations={citations}>
  {@html markedWithCiteExtension(content)}
</div>
```
```typescript
// enhance-citations-action.ts
export function enhanceCitations(node: HTMLElement, citations: Citation[]) {
  const citationMap = new Map(citations.map(c => [c.id, c]));
  
  function handleClick(e: MouseEvent) {
    const link = (e.target as HTMLElement)?.closest('[data-citation-id]');
    if (!link) return;
    
    e.preventDefault();
    const id = link.getAttribute('data-citation-id');
    const citation = citationMap.get(id);
    if (!citation) return;
    
    const urlParams = mapMetricsResolverQueryToUrlParams(citation.query);
    goto("?" + urlParams.toString());
  }
  
  node.addEventListener('click', handleClick);
  return { destroy() { node.removeEventListener('click', handleClick); } };
}
```

**Pros**: Minimal changes, keeps existing `marked` infrastructure

**Cons**: Citations are enhanced HTML, not true Svelte components (limited styling flexibility)

#### Phase 2: Component-Based Rendering (for icon buttons)

When we need more control over citation rendering (e.g., icon buttons per design requirements), migrate to a runtime Svelte markdown library like `svelte-exmarkdown` that can render true Svelte components inline.

```svelte
<!-- AssistantMessage.svelte with svelte-exmarkdown -->
<script>
  import Markdown from 'svelte-exmarkdown';
  import CitationLink from './CitationLink.svelte';
  import { citePlugin } from './cite-plugin';
  
  export let content: string;
  export let citations: Citation[];
</script>

<Markdown 
  md={content} 
  plugins={[citePlugin(citations)]}
/>
```
```svelte
<!-- CitationLink.svelte - true Svelte component -->
<script lang="ts">
  import { goto } from "$app/navigation";
  import { EmbedStore } from "../embeds/embed-store";
  import LinkIcon from "../icons/LinkIcon.svelte";
  
  export let text: string;
  export let query: MetricsResolverQuery;
  
  $: urlParams = mapMetricsResolverQueryToUrlParams(query);
  $: href = "?" + urlParams.toString();
  $: isEmbedded = EmbedStore.isEmbedded();
</script>

{#if isEmbedded}
  <button 
    type="button" 
    class="citation-link" 
    aria-label={text}
    title={text}
    on:click={() => goto(href)}
  >
    <LinkIcon />
  </button>
{:else}
  <a {href} data-sveltekit-preload-data="off" class="citation-link" title={text}>
    <LinkIcon />
  </a>
{/if}
```

**Pros**: True Svelte components, full styling flexibility (icon buttons, hover states, etc.)

**Cons**: New dependency, migration effort

Libraries to evaluate: `svelte-exmarkdown`, `@humanspeak/svelte-markdown` (both are mature, runtime-focused, support custom renderers)

## MCP Compatibility

This change is **fully backwards compatible** with external MCP clients:

| Client | What they use | Behavior |

|--------|--------------|----------|

| Claude Desktop | `open_url` from tool result | Links to Rill UI (external) |

| Cursor + Rill MCP | `open_url` from tool result | Links to Rill UI (external) |

| Rill Chat UI | Structured citations array | Native navigation, embed support |

The `open_url` field remains in the tool result - it's the universal interface. Rill's chat adds a convenience layer for its own frontend.

## Benefits

1. **Clean separation**: AI returns semantic data, frontend handles presentation
2. **No HTML parsing**: Citations are first-class data, not embedded in strings
3. **Extensible**: Easy to add hover previews, different click actions, etc.
4. **Frontend owns routing**: All URL/state knowledge stays in the frontend
5. **MCP compatible**: External clients continue working unchanged
6. **LLM-friendly syntax**: XML tags are familiar to LLMs and consistent with existing prompts
7. **Flexible rendering**: The `CitationLink` component can render citations as icon buttons (per design requirements) rather than inline text links - something that's difficult to achieve when citations are embedded in markdown strings

## Design Decisions

### 1. Citation Syntax: XML-style `<cite>` tags

```xml
<cite id="abc123">Revenue by country breakdown</cite>
```

**Why**:

- Consistent with existing prompt patterns (`<role>`, `<process>`, `<example>`)
- Avoids confusion with `{{ }}` Go template syntax
- LLMs handle XML/HTML well - less likely to malform than nested parentheses
- Multi-word labels work naturally
- Easy to parse with regex

### 2. Streaming Support

Streaming is the primary consumption method. In `CompleteStreaming`:

- Messages are sent to the client as they're created via `session.Subscribe()`
- Tool calls (`query_metrics_view`) arrive as individual messages during the completion

**Approach**: Include citations on the final `router_agent` result message.

```
Stream timeline:
  [query_metrics_view call] → [query_metrics_view result] → 
  [query_metrics_view call] → [query_metrics_view result] →
  [router_agent result with citations array]  ← citations included here
```

- Tool call messages stream as they happen (no change needed)
- When the final `router_agent` result is built, parse cite tags and build citations array
- Frontend doesn't need citations until rendering the final response anyway

## Migration Path

1. **Phase 1**: Proto and backend changes

   - Add `Citation` message and `citations` field to `Message` proto
   - Backend parses cite tags and builds citations array
   - Populate citations on `router_agent` result messages

2. **Phase 2**: Update AI prompt

   - Change analyst agent to output `<cite id="...">label</cite>` format
   - Update tool result to include tool call ID for reference

3. **Phase 3**: Frontend - Action-enhanced HTML

   - Implement custom `marked` extension for cite tags
   - Create Svelte action to hydrate citation links with navigation behavior
   - Remove old URL rewriting logic
   - Citations render as text links (same as today, but cleaner architecture)

4. **Phase 4**: Frontend - Component-based rendering (when icon buttons needed)

   - Evaluate and adopt `svelte-exmarkdown` or similar library
   - Create `CitationLink` Svelte component with icon button rendering
   - Migrate from action-enhanced HTML to true component rendering
   - Full design flexibility for citation appearance