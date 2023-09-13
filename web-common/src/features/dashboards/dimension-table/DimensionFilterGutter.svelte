<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import Circle from "@rilldata/web-common/components/icons/Circle.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import { CHECKMARK_COLORS } from "@rilldata/web-common/features/dashboards/config";

  import StickyHeader from "@rilldata/web-common/components/virtualized-table/core/StickyHeader.svelte";
  import { getContext } from "svelte";
  import type { VirtualizedTableConfig } from "../../../components/virtualized-table/types";
  import DimensionCompareMenu from "@rilldata/web-common/features/dashboards/leaderboard/DimensionCompareMenu.svelte";

  export let totalHeight: number;
  export let virtualRowItems;
  export let selectedIndex = [];
  export let excludeMode = false;
  export let isBeingCompared = false;

  function getColor(i) {
    const posInSelection = selectedIndex.indexOf(i);
    if (posInSelection >= 7) return "fill-gray-300";

    const colorIndex = posInSelection >= 0 ? posInSelection : i;
    return "fill-" + CHECKMARK_COLORS[colorIndex];
  }

  const config: VirtualizedTableConfig = getContext("config");
</script>

<div
  class="sticky left-0 top-0 z-20 bg-white"
  style:height="{totalHeight}px"
  style:width="{config.indexWidth}px"
>
  <div
    style:height="{config.columnHeaderHeight}px"
    class="sticky left-0 top-0 surface z-40 flex items-center"
  >
    <DimensionCompareMenu {isBeingCompared} on:toggle-dimension-comparison />
  </div>
  {#each virtualRowItems as row (`row-${row.key}`)}
    {@const isSelected = selectedIndex.includes(row.index)}
    <StickyHeader
      enableResize={false}
      position="left"
      header={{ size: config.indexWidth, start: row.start }}
    >
      <div class="py-0.5 grid place-items-center">
        {#if isSelected && isBeingCompared}
          <CheckCircle className={getColor(row.index)} size="18px" />
        {:else if isBeingCompared && !selectedIndex.length && row.index < 3}
          <Circle className={getColor(row.index)} size="16px" />
        {:else if isSelected && !excludeMode}
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
