<script lang="ts">
  import type { PivotDataRow } from "./types";
  import type { Row } from "@tanstack/svelte-table";
  import ChevronRight from "@rilldata/web-common/components/icons/ChevronRight.svelte";

  export let row: Row<PivotDataRow>;
  export let value: string;
  export let assembled = true;
</script>

<div
  class="flex gap-x-1"
  style:padding-left={`${row.depth * 2}rem`}
  class:font-normal={row.depth >= 1}
>
  {#if value === "LOADING_CELL"}
    <span class="loading-cell" />
  {:else if row.getCanExpand()}
    <button on:click={row.getToggleExpandedHandler()} class="cursor-pointer">
      <div class:rotate={row.getIsExpanded()} class="transition-transform">
        <ChevronRight />
      </div>
    </button>
  {/if}
  {value === "LOADING_CELL" ? "" : value}
</div>

<style lang="postcss">
  .loading-cell {
    @apply h-2 bg-gray-200 rounded w-full inline-block;
  }

  .rotate {
    @apply transform rotate-90;
  }
</style>
