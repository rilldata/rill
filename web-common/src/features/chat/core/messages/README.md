# Message Blocks

Transforms raw API messages (`V1Message`) into UI blocks (`Block`) for rendering chat conversations.

## Conceptual Model

```
V1Message (from API)          →  Block (for UI)
─────────────────────────────────────────────────────────
router_agent                  →  TextBlock (main conversation)
progress                      →  ThinkingBlock (grouped)
tool call (inline)            →  ThinkingBlock (grouped)
tool call (block)             →  ChartBlock / FileDiffBlock / etc.
tool call (hidden)            →  (not rendered)
result                        →  (attached to parent call)
(streaming, no text yet)      →  WorkingBlock (animated indicator)
```

## Key Abstractions

### `getBlockRoute()`

Centralizes all routing logic:

```typescript
function getBlockRoute(msg: V1Message): BlockRoute {
  // router_agent → "text" (main conversation)
  // progress → "thinking"
  // tool calls → consult registry (inline/block/hidden)
  // results → "skip" (attached to parent calls)
}
```

### Tool Registry

Configures how each tool renders:

- **`inline`** — Shown inside thinking blocks (most tools)
- **`block`** — Renders as a standalone block with its own header
- **`hidden`** — Not shown (internal orchestration)

Note: `router_agent` is NOT in the registry—it produces text, not thinking content.

## Transformation Flow

1. Build result map (tool call ID → result message)
2. For each message:
   - `getBlockRoute()` → text | thinking | block | skip
   - **text**: flush thinking buffer, add TextBlock
   - **thinking**: accumulate in buffer
   - **block**: flush thinking, add specific block (Chart, FileDiff, etc.)
   - **skip**: ignore (results are attached to their parent calls)
3. Flush any remaining thinking messages
4. Add WorkingBlock if streaming with no text response yet

## Adding New Block Types

1. Create block directory: `messages/my-block/`
2. Define type and factory: `my-block.ts`
3. Create component: `MyBlock.svelte`
4. Register in `tools/tool-registry.ts`:
   ```typescript
   [ToolName.MY_TOOL]: {
     renderMode: "block",
     createBlock: createMyBlock,
   },
   ```
5. Add to `Block` union in `block-transform.ts`
6. Add rendering case in `Messages.svelte`
