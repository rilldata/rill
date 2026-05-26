<script lang="ts">
  import EyeIcon from "@rilldata/web-common/components/icons/Eye.svelte";
  import EyeOffIcon from "@rilldata/web-common/components/icons/EyeInvisible.svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import type { DimensionTag, TagVisibilityState } from "./tag-utils";

  type Props = {
    tag: DimensionTag;
    visibility: TagVisibilityState;
    selected: boolean;
    onSelect: () => void;
    onShowAll: () => void;
    onHideAll: () => void;
    onShowOnly: () => void;
  };

  let {
    tag,
    visibility,
    selected,
    onSelect,
    onShowAll,
    onHideAll,
    onShowOnly,
  }: Props = $props();

  const actionBtnClass =
    "flex items-center justify-center h-[22px] w-[22px] rounded-sm text-icon-muted hover:text-fg-primary hover:bg-surface-background transition-colors";
</script>

<div
  class="group w-full flex items-center gap-x-1 px-2 py-1 rounded-sm"
  class:bg-popover-accent={selected}
  class:hover:bg-surface-hover={!selected}
>
  <button
    type="button"
    class="flex items-center gap-x-2 flex-1 min-w-0 text-left"
    onclick={onSelect}
    aria-label={`${selected ? "Clear filter" : "Filter by"} ${tag.name}`}
    aria-pressed={selected}
  >
    <svg width="12" height="12" viewBox="0 0 16 16" aria-hidden="true">
      {#if visibility.state === "all"}
        <circle cx="8" cy="8" r="6" class="fill-theme-500" />
      {:else if visibility.state === "partial"}
        <circle
          cx="8"
          cy="8"
          r="6"
          class="fill-none stroke-theme-500"
          stroke-width="1.5"
        />
        <path d="M8 2 A 6 6 0 0 1 8 14 Z" class="fill-theme-500" />
      {:else}
        <circle
          cx="8"
          cy="8"
          r="6"
          class="fill-none stroke-fg-secondary"
          stroke-width="1.5"
        />
      {/if}
    </svg>
    <span class="truncate flex-1 min-w-0 text-fg-primary text-sm">
      {tag.name}
    </span>
    <span
      class="tabular-nums text-xs text-fg-secondary flex-none"
      aria-label={`${visibility.visibleCount} of ${visibility.totalCount} shown`}
    >
      {visibility.visibleCount}/{visibility.totalCount}
    </span>
  </button>

  <div
    class="flex items-center gap-x-0.5 flex-none opacity-0 group-hover:opacity-100 transition-opacity"
    class:opacity-100={selected}
  >
    <Tooltip.Root delayDuration={200}>
      <Tooltip.Trigger>
        <button
          type="button"
          class={actionBtnClass}
          onclick={(e) => {
            e.stopPropagation();
            onShowOnly();
          }}
          aria-label={`Only show ${tag.name}`}
        >
          <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
            <circle cx="8" cy="8" r="4.5" class="fill-current" />
            <circle
              cx="8"
              cy="8"
              r="7"
              class="stroke-current fill-none"
              stroke-width="1.25"
            />
          </svg>
        </button>
      </Tooltip.Trigger>
      <Tooltip.Content
        side="top"
        class="bg-popover text-fg-primary z-popover text-xs px-2 py-1"
      >
        Only show this tag
      </Tooltip.Content>
    </Tooltip.Root>

    <Tooltip.Root delayDuration={200}>
      <Tooltip.Trigger>
        <button
          type="button"
          class={actionBtnClass}
          onclick={(e) => {
            e.stopPropagation();
            onShowAll();
          }}
          aria-label={`Show all in ${tag.name}`}
        >
          <EyeIcon size="14px" color="currentColor" />
        </button>
      </Tooltip.Trigger>
      <Tooltip.Content
        side="top"
        class="bg-popover text-fg-primary z-popover text-xs px-2 py-1"
      >
        Show all in tag
      </Tooltip.Content>
    </Tooltip.Root>

    <Tooltip.Root delayDuration={200}>
      <Tooltip.Trigger>
        <button
          type="button"
          class={actionBtnClass}
          onclick={(e) => {
            e.stopPropagation();
            onHideAll();
          }}
          aria-label={`Hide all in ${tag.name}`}
        >
          <EyeOffIcon size="14px" color="currentColor" />
        </button>
      </Tooltip.Trigger>
      <Tooltip.Content
        side="top"
        class="bg-popover text-fg-primary z-popover text-xs px-2 py-1"
      >
        Hide all in tag
      </Tooltip.Content>
    </Tooltip.Root>
  </div>
</div>
