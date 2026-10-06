<!--
  "Memory updated" notice rendered under an assistant turn whenever the AI wrote to the user's memory,
  either explicitly (update_memory) or in the background (extract_memories).
  Lists what changed and offers a one-click undo that reverses every operation in the notice.
-->
<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { Brain, ChevronDown, ChevronRight } from "lucide-svelte";
  import MemoryDialog from "./MemoryDialog.svelte";
  import { MemoryStatus, updateAIMemory, useAIMemories } from "./memory-store";
  import type { MemoryOp, MemoryUpdateBlock } from "./memory-update-block";

  export let block: MemoryUpdateBlock;

  const client = useRuntimeClient();
  const memoriesQuery = useAIMemories(client, true);

  let expanded = false;
  let undoing = false;
  let dialogOpen = false;

  $: memoriesById = new Map(
    ($memoriesQuery.data?.memories ?? []).map((mem) => [mem.id, mem]),
  );

  // The notice is "undone" when every operation has been reversed, whether through the Undo button here
  // or through the memory manager. Deriving it from the server keeps the state correct across reloads.
  $: undone =
    $memoriesQuery.isSuccess &&
    !!$memoriesQuery.data?.enabled &&
    block.ops.length > 0 &&
    block.ops.every((op) => isReversed(op));

  function isReversed(op: MemoryOp): boolean {
    const memory = memoriesById.get(op.memory_id);
    switch (op.op) {
      case "add":
        return !memory || memory.status !== MemoryStatus.ACTIVE;
      case "delete":
        return !!memory && memory.status === MemoryStatus.ACTIVE;
      case "update":
        return (
          !memory ||
          (memory.content === op.previous_content &&
            memory.category === op.previous_category)
        );
      default:
        return true;
    }
  }

  async function undo() {
    undoing = true;
    try {
      for (const op of block.ops) {
        if (!op.memory_id || isReversed(op)) continue;
        switch (op.op) {
          case "add":
            await updateAIMemory(client, op.memory_id, {
              status: MemoryStatus.DELETED,
            });
            break;
          case "delete":
            await updateAIMemory(client, op.memory_id, {
              status: MemoryStatus.ACTIVE,
            });
            break;
          case "update":
            // An update can change both fields, so both are restored.
            await updateAIMemory(client, op.memory_id, {
              content: op.previous_content,
              category: op.previous_category,
            });
            break;
        }
      }
    } catch (err) {
      console.error("Failed to undo memory change", err);
      eventBus.emit("notification", {
        message: m.chat_memory_undo_failed(),
        type: "error",
      });
    } finally {
      undoing = false;
    }
  }

  function opLabel(op: MemoryOp): string {
    switch (op.op) {
      case "add":
        return m.chat_memory_op_added();
      case "update":
        return m.chat_memory_op_updated();
      case "delete":
        return m.chat_memory_op_removed();
      default:
        return "";
    }
  }
</script>

<div class="memory-chip" class:undone data-testid="memory-update-chip">
  <button
    type="button"
    class="memory-chip-header"
    aria-expanded={expanded}
    onclick={() => (expanded = !expanded)}
  >
    {#if expanded}
      <ChevronDown size="14px" />
    {:else}
      <ChevronRight size="14px" />
    {/if}
    <Brain size="14px" />
    <span class="memory-chip-title">
      {undone ? m.chat_memory_undone() : m.chat_memory_updated()}
    </span>
    {#if !expanded}
      <span class="memory-chip-preview">{block.ops[0]?.content}</span>
    {/if}
  </button>
  <div class="memory-chip-actions">
    {#if !undone}
      <Button type="text" compact disabled={undoing} onClick={undo}>
        {m.chat_memory_undo()}
      </Button>
    {/if}
    <Button type="text" compact onClick={() => (dialogOpen = true)}>
      {m.chat_memory_open()}
    </Button>
  </div>
  {#if expanded}
    <ul class="memory-chip-ops">
      {#each block.ops as op (op.memory_id + op.op)}
        <li>
          <span class="memory-op-label" data-op={op.op}>{opLabel(op)}</span>
          <span class="memory-op-content">{op.content}</span>
          {#if op.op === "update" && op.previous_content}
            <span class="memory-op-previous">
              {m.chat_memory_op_previously({ content: op.previous_content })}
            </span>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</div>

<MemoryDialog bind:open={dialogOpen} />

<style lang="postcss">
  .memory-chip {
    @apply w-full max-w-full self-start;
    @apply grid items-center gap-x-2;
    @apply rounded-md border border-border bg-surface-subtle;
    @apply px-2 py-1 text-xs text-fg-secondary;
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .memory-chip.undone {
    @apply opacity-70;
  }

  .memory-chip-header {
    @apply flex items-center gap-x-1.5 min-w-0;
    @apply bg-transparent border-none p-0 text-left cursor-pointer;
    color: inherit;
    font: inherit;
  }

  .memory-chip-title {
    @apply font-medium text-fg-primary whitespace-nowrap;
  }

  .memory-chip-preview {
    @apply truncate text-fg-muted;
  }

  .memory-chip-actions {
    @apply flex items-center gap-x-1 shrink-0;
  }

  .memory-chip-ops {
    @apply col-span-2 list-none m-0 mt-1 p-0 pl-5 flex flex-col gap-y-1;
  }

  .memory-chip-ops li {
    @apply flex flex-wrap items-baseline gap-x-2;
  }

  .memory-op-label {
    @apply uppercase tracking-wide text-[10px] font-semibold;
  }

  .memory-op-label[data-op="add"] {
    @apply text-primary-600;
  }

  .memory-op-label[data-op="update"] {
    @apply text-fg-secondary;
  }

  .memory-op-label[data-op="delete"] {
    @apply text-fg-muted;
  }

  .memory-op-content {
    @apply text-fg-primary;
  }

  .memory-op-previous {
    @apply w-full text-fg-muted line-through;
  }
</style>
