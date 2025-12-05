# Message Blocks

This directory transforms raw API messages (`V1Message`) into UI blocks (`Block`) for rendering chat conversations.

## Conceptual Model

### Message Sources → UI Blocks

```
V1Message (from API)          →  Block (for UI)
─────────────────────────────────────────────────────────
router_agent                  →  TextBlock (main conversation)
progress                      →  ThinkingBlock (grouped)
tool call (inline)            →  ThinkingBlock (grouped)
tool call (block)             →  ThinkingBlock + ChartBlock/etc.
tool call (hidden)            →  (not rendered)
result                        →  (attached to parent call)
(streaming, no response yet)  →  PlanningBlock (placeholder)
```

### Block Flow

```
                    ┌─────────────────┐
                    │  PlanningBlock  │  ← "AI is working" (before response arrives)
                    └─────────────────┘
                            │
            ┌───────────────┴───────────────┐
            ↓                               ↓
    (AI uses tools)                  (AI responds directly)
    ┌───────────────┐                       │
    │ ThinkingBlock │                       │
    └───────────────┘                       │
            ↓                               ↓
    ┌───────────────┐               ┌───────────────┐
    │   TextBlock   │               │   TextBlock   │
    └───────────────┘               └───────────────┘
```

### Key Abstraction: `getBlockRoute()`

All routing logic is centralized in one function:

```typescript
function getBlockRoute(msg: V1Message): BlockRoute {
  // router_agent → "text" (main conversation)
  // progress → "thinking"
  // tool calls → consult registry (inline/block/hidden)
  // results → "skip" (attached to parent calls)
}
```

This makes the transformation loop trivial—just a switch on the route.

### Tool Registry

The **tool registry** (`tool-registry.ts`) configures how tool calls render:

- **`inline`** - Shown in thinking blocks (most tools)
- **`block`** - Shown in thinking, then produces a top-level block
- **`hidden`** - Not shown (internal orchestration agents)

Note: `router_agent` is NOT in the registry—it produces TEXT, not thinking content.

## Directory Structure

Each block type has its own directory:

```
messages/
├── block-transform.ts           # Transformation: V1Message → Block
├── tool-registry.ts             # Tool rendering configuration
├── Messages.svelte              # Main container component
│
├── ShimmerText.svelte           # Shared: loading animation
├── Error.svelte                 # Shared: error display
│
├── text/                        # Main conversation (user/assistant)
│   ├── text-block.ts
│   ├── AssistantMessage.svelte
│   ├── UserMessage.svelte
│   └── rewrite-citation-urls.ts
│
├── planning/                    # "AI is working" placeholder
│   ├── planning-block.ts        # Type + shouldShowPlanning()
│   └── PlanningBlock.svelte
│
├── thinking/                    # AI reasoning (progress + tool calls)
│   ├── thinking-block.ts
│   ├── ThinkingBlock.svelte
│   ├── CallMessage.svelte
│   ├── tool-display-names.ts
│   └── tool-icons.ts
│
├── chart/                       # Chart visualizations
│   ├── chart-block.ts
│   └── ChartBlock.svelte
│
└── file-diff/                   # File change diffs
    ├── file-diff-block.ts
    └── FileDiffBlock.svelte
```

## Transformation Flow

```
1. Build result map (tool call ID → result message)
2. For each message:
   - getBlockRoute() → text | thinking | block | skip
   - text: flush thinking, add TextBlock
   - thinking: accumulate in buffer
   - block: accumulate, flush thinking, add block
   - skip: ignore
3. Flush remaining thinking
4. Add planning indicator if streaming with no response
```

## Adding New Block Types

To add a new block-level tool (like `ChartBlock` or `FileDiffBlock`):

1. **Create block directory**: `messages/my-block/`
2. **Define type and factory**: `my-block.ts`
3. **Create component**: `MyBlock.svelte`
4. **Register in tool registry**:
   ```typescript
   [ToolName.MY_TOOL]: {
     renderMode: "block",
     createBlock: createMyBlock,
   },
   ```
5. **Add to Block union** in `block-transform.ts`
6. **Add rendering case** in `Messages.svelte`

No changes to transformation logic needed—the registry handles routing.
