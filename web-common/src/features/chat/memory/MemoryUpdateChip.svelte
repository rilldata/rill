<!--
  "Memory updated" notice rendered under an assistant turn whenever the AI wrote to the user's memory,
  either explicitly (update_memory) or in the background (extract_memories).
  Lists what changed and links to the memory manager, where any change can be edited or deleted.
-->
<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { Brain, ChevronDown, ChevronRight } from "lucide-svelte";
  import MemoryDialog from "./MemoryDialog.svelte";
  import type { MemoryOp, MemoryUpdateBlock } from "./memory-update-block";

  export let block: MemoryUpdateBlock;

  let expanded = false;
  let dialogOpen = false;

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

<div class="memory-chip" data-testid="memory-update-chip">
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
    <span class="memory-chip-title">{m.chat_memory_updated()}</span>
    {#if !expanded}
      <span class="memory-chip-preview">{block.ops[0]?.content}</span>
    {/if}
  </button>
  <div class="memory-chip-actions">
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
</style>
