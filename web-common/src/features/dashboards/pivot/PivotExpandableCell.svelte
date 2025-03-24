<script lang="ts">
  import ChevronRight from "@rilldata/web-common/components/icons/ChevronRight.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import type { Row } from "@tanstack/svelte-table";
  import type { PivotDataRow } from "./types";

  export let row: Row<PivotDataRow>;
  export let value: string;
  export let assembled = true;

  function handleClick() {
    if (row.getCanExpand()) {
      row.getToggleExpandedHandler()();
    }
  }
</script>

<div
  role="presentation"
  class="dimension-cell"
  style:padding-left={`${row.depth * 14}px`}
  class:-ml-1={assembled && row.getCanExpand()}
  class:cursor-pointer={assembled && row.getCanExpand()}
  on:click|stopPropagation={handleClick}
>
  {#if value === "LOADING_CELL"}
    <span class="loading-cell" />
  {:else if assembled && row.getCanExpand()}
    <div class="caret" class:expanded={row.getIsExpanded()}>
      <div class:rotate={row.getIsExpanded()}>
        <ChevronRight size="16px" color="#9CA3AF" />
      </div>
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

  .rotate {
    @apply transform rotate-90;
  }

  .dimension-cell {
    @apply flex gap-x-0.5;
  }

  .caret {
    @apply opacity-0;
  }
  .dimension-cell:hover .caret {
    @apply opacity-100;
  }

  .caret.expanded {
    @apply opacity-100;
  }
</style>
