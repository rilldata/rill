<!--
  "Memory updated" notice rendered under an assistant turn when the AI wrote to the user's memory through update_memory.
  Shows what changed and links to the memory manager, where the change can be edited or deleted.
-->
<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { Brain } from "lucide-svelte";
  import MemoryDialog from "./MemoryDialog.svelte";
  import type { MemoryUpdateBlock } from "./memory-update-block";

  export let block: MemoryUpdateBlock;

  let dialogOpen = false;

  $: opLabel = labelFor(block.result.op);

  function labelFor(op: string): string {
    switch (op) {
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
  <Brain size="14px" class="shrink-0" />
  <span class="memory-chip-title">{m.chat_memory_updated()}</span>
  <span class="memory-op-label" data-op={block.result.op}>{opLabel}</span>
  <span class="memory-op-content" title={block.result.content}>
    {block.result.content}
  </span>
  <div class="memory-chip-actions">
    <Button type="text" compact onClick={() => (dialogOpen = true)}>
      {m.chat_memory_open()}
    </Button>
  </div>
</div>

<MemoryDialog bind:open={dialogOpen} />

<style lang="postcss">
  .memory-chip {
    @apply w-full max-w-full self-start;
    @apply flex items-center gap-x-2 min-w-0;
    @apply rounded-md border border-border bg-surface-subtle;
    @apply px-2 py-1 text-xs text-fg-secondary;
  }

  .memory-chip-title {
    @apply font-medium text-fg-primary whitespace-nowrap;
  }

  .memory-op-label {
    @apply uppercase tracking-wide text-[10px] font-semibold whitespace-nowrap;
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
    @apply flex-1 min-w-0 truncate text-fg-primary;
  }

  .memory-chip-actions {
    @apply flex items-center shrink-0;
  }
</style>
