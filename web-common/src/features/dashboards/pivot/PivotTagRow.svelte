<script lang="ts">
  import type { DimensionTag } from "@rilldata/web-common/components/menu/tag-utils";
  import Column from "@rilldata/web-common/components/icons/Column.svelte";
  import Pivot from "@rilldata/web-common/components/icons/Pivot.svelte";
  import Row from "@rilldata/web-common/components/icons/Row.svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { modifierHeld } from "@rilldata/web-common/lib/modifier-key";
  import type { PivotChipData } from "./types";

  type Props = {
    tag: DimensionTag;
    dimensions: PivotChipData[];
    measures: PivotChipData[];
    selected: boolean;
    onAddRows: (replace: boolean) => void;
    onAddColumns: (replace: boolean) => void;
    onAutoArrange: (replace: boolean) => void;
    onDragStart: (event: MouseEvent, rect: DOMRect) => void;
  };

  let {
    tag,
    dimensions,
    measures,
    selected,
    onAddRows,
    onAddColumns,
    onAutoArrange,
    onDragStart,
  }: Props = $props();

  let rowEl: HTMLDivElement | undefined = $state();

  const dimensionCount = $derived(dimensions.length);
  const measureCount = $derived(measures.length);

  const actionBtnClass =
    "flex items-center justify-center h-[18px] w-[18px] rounded-sm text-icon-muted hover:text-fg-primary hover:bg-surface-background transition-colors";

  function handleMouseDown(e: MouseEvent) {
    if (e.button !== 0) return;
    const target = e.target as HTMLElement | null;
    if (target?.closest("button")) return;
    if (!rowEl) return;
    e.preventDefault();
    onDragStart(e, rowEl.getBoundingClientRect());
  }

  function handleClick(
    e: MouseEvent,
    action: (replace: boolean) => void,
  ) {
    // Read the modifier off the event itself so a click that happens before
    // the global keydown fires still picks up the held key.
    action(e.metaKey || e.ctrlKey);
  }
</script>

<div
  bind:this={rowEl}
  class="tag-row group"
  class:selected
  role="presentation"
  onmousedown={handleMouseDown}
>
  <span class="truncate flex-1 min-w-0 text-fg-primary">
    {tag.name}
  </span>

  <div class="flex items-center gap-x-1 flex-none group-hover:hidden">
    {#if measureCount > 0}
      <span class="count-tile meas-tile" title={`${measureCount} measures`}>
        {measureCount}
      </span>
    {/if}
    {#if dimensionCount > 0}
      <span class="count-tile dim-tile" title={`${dimensionCount} dimensions`}>
        {dimensionCount}
      </span>
    {/if}
  </div>

  <div class="hidden group-hover:flex items-center gap-x-0.5 flex-none">
    {#if dimensionCount > 0}
      <Tooltip.Root delayDuration={200}>
        <Tooltip.Trigger>
          <button
            type="button"
            class={actionBtnClass}
            onclick={(e) => handleClick(e, onAddRows)}
            aria-label={$modifierHeld
              ? `Replace rows with dimensions in ${tag.name}`
              : `Add all dimensions in ${tag.name} to rows`}
          >
            <Row size="14px" color="currentColor" />
          </button>
        </Tooltip.Trigger>
        <Tooltip.Content
          side="top"
          class="bg-popover text-fg-primary z-popover text-xs px-2 py-1"
        >
          {#if $modifierHeld}
            Replace rows with this tag's dimensions
          {:else}
            <div>Add all to rows</div>
            <div class="hint">
              <span class="kbd">⌘</span> + Click to replace
            </div>
          {/if}
        </Tooltip.Content>
      </Tooltip.Root>
    {/if}

    <Tooltip.Root delayDuration={200}>
      <Tooltip.Trigger>
        <button
          type="button"
          class={actionBtnClass}
          onclick={(e) => handleClick(e, onAddColumns)}
          aria-label={$modifierHeld
            ? `Replace columns with items in ${tag.name}`
            : `Add all in ${tag.name} to columns`}
        >
          <Column size="14px" color="currentColor" />
        </button>
      </Tooltip.Trigger>
      <Tooltip.Content
        side="top"
        class="bg-popover text-fg-primary z-popover text-xs px-2 py-1"
      >
        {#if $modifierHeld}
          Replace columns with this tag's items
        {:else}
          <div>Add all to columns</div>
          <div class="hint">
            <span class="kbd">⌘</span> + Click to replace
          </div>
        {/if}
      </Tooltip.Content>
    </Tooltip.Root>

    {#if dimensionCount > 0 && measureCount > 0}
      <Tooltip.Root delayDuration={200}>
        <Tooltip.Trigger>
          <button
            type="button"
            class={actionBtnClass}
            onclick={(e) => handleClick(e, onAutoArrange)}
            aria-label={$modifierHeld
              ? `Replace rows and columns with auto-arranged ${tag.name}`
              : `Auto-arrange ${tag.name}: dimensions to rows, measures to columns`}
          >
            <Pivot size="14px" color="currentColor" />
          </button>
        </Tooltip.Trigger>
        <Tooltip.Content
          side="top"
          class="bg-popover text-fg-primary z-popover text-xs px-2 py-1"
        >
          {#if $modifierHeld}
            Replace rows and columns with this tag
          {:else}
            <div>Auto-arrange</div>
            <div class="hint">
              <span class="kbd">⌘</span> + Click to replace
            </div>
          {/if}
        </Tooltip.Content>
      </Tooltip.Root>
    {/if}
  </div>
</div>

<style lang="postcss">
  .tag-row {
    @apply w-full flex items-center gap-x-1 px-1.5 py-0.5 rounded-sm;
    @apply cursor-grab select-none;
    @apply hover:bg-surface-hover;
  }

  .tag-row.selected {
    @apply bg-popover-accent;
  }

  .tag-row.selected:hover {
    @apply bg-popover-accent;
  }

  .tag-row:active {
    @apply cursor-grabbing;
  }

  .count-tile {
    @apply inline-flex items-center justify-center;
    @apply tabular-nums text-[10px] font-medium;
    @apply min-w-[16px] h-[16px] px-1 rounded-sm border;
  }

  .dim-tile {
    @apply bg-theme-50 border-theme-200 text-theme-800;
  }

  .meas-tile {
    @apply bg-theme-secondary-50 border-theme-secondary-200 text-theme-secondary-800;
  }

  .kbd {
    @apply inline-block px-1 py-px rounded-sm border;
    @apply text-[10px] font-mono text-fg-secondary;
  }

  .hint {
    @apply mt-0.5 text-[10px] text-fg-secondary;
  }
</style>
