<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import CancelCircle from "../icons/CancelCircle.svelte";

  type Props = {
    tagName: string;
    onClear: () => void;
  };

  let { tagName, onClear }: Props = $props();
</script>

<div class="flex items-center gap-x-1.5 px-3 py-1.5 bg-popover-accent border-b">
  <span class="text-xs text-fg-secondary flex-none">Filter:</span>
  <span class="truncate text-xs text-fg-primary font-medium flex-1 min-w-0">
    {tagName}
  </span>
  <Tooltip.Root delayDuration={200}>
    <Tooltip.Trigger>
      {#snippet child({ props })}
        <button
          {...props}
          type="button"
          class="flex-none text-icon-muted hover:text-fg-primary transition-colors"
          onclick={onClear}
          aria-label="Clear tag filter"
        >
          <CancelCircle size="14px" />
        </button>
      {/snippet}
    </Tooltip.Trigger>
    <Tooltip.Content
      side="top"
      class="bg-popover text-fg-primary z-popover text-xs px-2 py-1"
    >
      Clear filter
    </Tooltip.Content>
  </Tooltip.Root>
</div>
