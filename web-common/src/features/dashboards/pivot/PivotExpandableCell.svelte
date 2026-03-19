<script lang="ts">
  import ChevronRight from "@rilldata/web-common/components/icons/ChevronRight.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import { LOADING_CELL } from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
  import type { Row } from "@tanstack/svelte-table";
  import type { PivotDataRow } from "./types";

  export let row: Row<PivotDataRow>;
  export let value: string;
  export let assembled = true;
  export let hasNestedDimensions = false;

  $: canExpand = row.getCanExpand();
  $: expanded = row.getIsExpanded();
  $: assembledAndCanExpand = assembled && canExpand;

  $: needsSpacer = row.depth >= 1 || (hasNestedDimensions && !canExpand);

  function handleExpandClick(e: MouseEvent) {
    e.stopPropagation();
    if (assembledAndCanExpand) {
      row.getToggleExpandedHandler()();
    }
  }
</script>

<div
  role="presentation"
  class="dimension-cell"
  style:padding-left="{row.depth * 14}px"
>
  {#if value === LOADING_CELL}
    <span class="loading-cell" />
  {:else if assembledAndCanExpand}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <div
      role="button"
      tabindex="-1"
      class="caret opacity-100 shrink-0 cursor-pointer"
      class:expanded
      on:click={handleExpandClick}
    >
      <ChevronRight size="16px" color="#9CA3AF" />
    </div>
  {:else if needsSpacer}
    <span class="shrink-0"><Spacer size="16px" /></span>
  {/if}

  <span class="truncate min-w-0">
    {#if value === LOADING_CELL}
      {""}
    {:else if value === ""}
      {"\u00A0"}
    {:else}
      {value}
    {/if}
  </span>
</div>

<style lang="postcss">
  .loading-cell {
    @apply h-2 bg-gray-200 rounded w-full inline-block;
  }

  .dimension-cell {
    @apply flex gap-x-0.5;
  }

  .caret.expanded {
    @apply opacity-100 transform rotate-90;
  }
</style>
