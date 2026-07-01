<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { detectOverflow } from "@rilldata/web-common/lib/actions/detect-overflow";

  type Props = {
    name: string;
    count: number;
    selected: boolean;
    onSelect: () => void;
  };

  let { name, count, selected, onSelect }: Props = $props();

  // Only show the tooltip when the tag name is actually clipped, matching the
  // pivot tag rows.
  let isTruncated = $state(false);
</script>

<button class="tag-row" class:selected onclick={onSelect}>
  <Tooltip.Root delayDuration={200} disabled={!isTruncated}>
    <Tooltip.Trigger>
      {#snippet child({ props })}
        <span
          {...props}
          class="truncate flex-1 min-w-0 text-left text-fg-primary"
          use:detectOverflow={(v) => (isTruncated = v)}
        >
          {name}
        </span>
      {/snippet}
    </Tooltip.Trigger>
    <Tooltip.Content
      side="top"
      class="bg-popover text-fg-primary z-popover text-xs px-2 py-1"
    >
      {name}
    </Tooltip.Content>
  </Tooltip.Root>

  <span class="count-tile" title={`${count} dashboards`}>{count}</span>
</button>

<style lang="postcss">
  .tag-row {
    @apply w-full flex items-center gap-x-1 px-1.5 py-1 rounded-sm;
    @apply text-sm cursor-pointer;
    @apply hover:bg-surface-hover;
  }

  .tag-row.selected,
  .tag-row.selected:hover {
    @apply bg-popover-accent;
  }

  .count-tile {
    @apply inline-flex items-center justify-center flex-none;
    @apply tabular-nums text-[10px] font-medium text-fg-secondary;
    @apply min-w-[16px] h-[16px] px-1 rounded-sm border;
  }
</style>
