<script lang="ts">
  import ChevronRight from "@rilldata/web-common/components/icons/ChevronRight.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import type { Row } from "@tanstack/svelte-table";
  import type { PivotDataRow } from "./types";

  export let row: Row<PivotDataRow>;
  export let value: string;
  export let assembled = true;

  $: canExpand = row.getCanExpand();
  $: expanded = row.getIsExpanded();
  $: assembledAndCanExpand = assembled && canExpand;
</script>

<div
  role="presentation"
  class="dimension-cell pointer-events-none"
  style:padding-left="{row.depth * 14}px"
  class:-ml-1={assembledAndCanExpand}
  class:cursor-pointer={assembledAndCanExpand}
>
  {#if value === "LOADING_CELL"}
    <span class="loading-cell" />
  {:else if assembledAndCanExpand}
    <div class="caret opacity-100" class:expanded>
      <ChevronRight size="16px" color="#9CA3AF" />
    </div>
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

  .dimension-cell {
    @apply flex gap-x-0.5;
  }

  .caret.expanded {
    @apply opacity-100 transform rotate-90;
  }
</style>
