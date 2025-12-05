# Message Blocks

This directory contains the presentation layer for chat conversations. It transforms raw API messages into structured blocks for rendering.

## Architecture

### Message Block Types

Messages are transformed into one of three block types for display:

- **`TextMessage`** - A single text message (user prompt or assistant response)
- **`ThinkingBlock`** - A collapsible block containing grouped progress updates and tool calls
- **`ChartBlock`** - A visualization created from a `create_chart` tool call

These are unified under the `MessageBlock` discriminated union for rendering purposes.

### Directory Structure

Each block type is **completely self-contained** in its own directory:

```
messages/
├── text/
│   ├── text-message.ts          # TextMessage type & creation
│   ├── TextMessage.svelte        # Rendering component
│   └── rewrite-citation-urls.ts  # Citation URL utilities
│
├── thinking/
│   ├── thinking-block.ts         # ThinkingBlock type & utilities
│   ├── ThinkingBlock.svelte      # Main collapsible component
│   ├── CallMessage.svelte        # Tool call display
│   ├── ProgressMessage.svelte    # Progress update display
│   ├── ShimmerText.svelte        # Loading animation
│   ├── tool-display-names.ts    # Tool metadata
│   └── tool-icons.ts             # Tool icon mappings
│
├── chart/
│   ├── chart-block.ts            # ChartBlock type, parsing & creation
│   └── ChartBlock.svelte         # Chart rendering
│
├── message-blocks.ts             # MessageBlockTransformer (orchestration)
├── Messages.svelte               # Main container component
└── Error.svelte                  # Shared error display
```

## Key Concepts

### Transformation Pipeline

1. **Raw messages** (from API) → `Conversation.getMessageBlocks()`
2. **Message filtering** → Filter out internal tool results
3. **Grouping logic** → `MessageBlockTransformer.transform()`
   - Text messages remain standalone
   - Progress/tool call messages are grouped into thinking blocks
   - `create_chart` tool calls **end** the current thinking block (so charts appear at top level)
   - Planning indicator added when waiting for AI response
4. **Message blocks** → Rendered by `Messages.svelte`

### Thinking Block Splitting

Thinking blocks are split on `create_chart` tool calls to achieve this flow:

```
Thinking Block (progress + create_chart call)
  ↓
Chart Visualization
  ↓
Thinking Block (new progress + create_chart call)
  ↓
Chart Visualization
```

This ensures each chart appears prominently between thinking blocks rather than hidden inside them.

### Planning Indicator

When the user sends a message but no AI response has arrived yet, a placeholder thinking block is shown with "Thinking..." to provide immediate feedback.

## Adding New Block Types

To add a new block type:

1. Create a new directory: `messages/new-type/`
2. Add type definition and creation logic: `new-type.ts`
3. Add rendering component: `NewType.svelte`
4. Update `message-blocks.ts` to handle the new type
5. Update `Messages.svelte` to render it

## Testing

When modifying block transformation logic, verify:
- Charts split thinking blocks correctly
- Planning indicator appears/disappears smoothly
- Empty thinking blocks are hidden
- All message types render in correct order


