<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import StickyHeader from "@rilldata/web-common/components/virtualized-table/core/StickyHeader.svelte";
  import { getContext } from "svelte";
  import type { VirtualizedTableConfig } from "../../../components/virtualized-table/types";

  export let totalHeight: number;
  export let virtualRowItems;
  export let selectedIndex = [];
  export let excludeMode = false;

  const config: VirtualizedTableConfig = getContext("config");
</script>

<div
  class="sticky left-0 top-0 z-20"
  style:height="{totalHeight}px"
  style:width="{config.indexWidth}px"
>
  <!-- Hide filter symbols above the column headers -->
  <div
    style:height="{config.columnHeaderHeight}px"
    class="sticky left-0 top-0 surface z-40"
  />
  {#each virtualRowItems as row (`row-${row.key}`)}
    {@const isSelected = selectedIndex.includes(row.index)}
    <StickyHeader
      enableResize={false}
      position="left"
      header={{ size: config.indexWidth, start: row.start }}
    >
      <div class="py-0.5 grid place-items-center">
        {#if isSelected && !excludeMode}
          <Check size="20px" />
        {:else if isSelected && excludeMode}
          <Cancel size="20px" />
        {:else}
          <Spacer />
        {/if}
      </div>
    </StickyHeader>
  {/each}
</div>
