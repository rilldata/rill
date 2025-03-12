<script lang="ts">
  import ChevronRight from "@rilldata/web-common/components/icons/ChevronRight.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import type { Row } from "@tanstack/svelte-table";
  import type { PivotDataRow } from "./types";

  export let row: Row<PivotDataRow>;
  export let value: string;
  export let assembled = true;
</script>

<div class="flex gap-x-1" style:padding-left={`${row.depth * 14}px`}>
  {#if value === "LOADING_CELL"}
    <span class="loading-cell" />
  {:else if assembled && row.getCanExpand()}
    <button
      on:click|stopPropagation={row.getToggleExpandedHandler()}
      class="cursor-pointer px-0.5 -m-1 pointer-events-auto"
    >
      <div class:rotate={row.getIsExpanded()} class="transition-transform">
        <ChevronRight size="16px" color="#9CA3AF" />
      </div>
    </button>
  {:else if row.depth >= 1}
    <Spacer size="16px" />
  {/if}

  <span class="truncate">
    {#if value === "LOADING_CELL"}
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

  .rotate {
    @apply transform rotate-90;
  }
</style>
