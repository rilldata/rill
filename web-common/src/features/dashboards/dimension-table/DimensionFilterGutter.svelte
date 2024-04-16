<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import {
    COMPARIONS_COLORS,
    SELECTED_NOT_COMPARED_COLOR,
  } from "@rilldata/web-common/features/dashboards/config";

  import StickyHeader from "@rilldata/web-common/components/virtualized-table/core/StickyHeader.svelte";
  import DimensionCompareMenu from "@rilldata/web-common/features/dashboards/leaderboard/DimensionCompareMenu.svelte";
  import { getContext } from "svelte";
  import type { VirtualizedTableConfig } from "../../../components/virtualized-table/types";
  import { getStateManagers } from "../state-managers/state-managers";

  export let totalHeight: number;
  export let virtualRowItems;
  export let selectedIndex: number[] = [];
  export let excludeMode = false;
  export let isBeingCompared = false;

  const {
    selectors: {
      dimensions: { dimensionTableDimName },
    },
  } = getStateManagers();

  function getColor(i) {
    const posInSelection = selectedIndex.indexOf(i);
    if (posInSelection >= 7) return SELECTED_NOT_COMPARED_COLOR;
    return COMPARIONS_COLORS[posInSelection];
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
    <DimensionCompareMenu dimensionName={$dimensionTableDimName} />
  </div>
  {#each virtualRowItems as row (`row-${row.key}`)}
    {@const isSelected = selectedIndex.includes(row.index)}
    <StickyHeader
      enableResize={false}
      position="left"
      header={{ size: config.indexWidth, start: row.start }}
    >
      <div class="py-0.5 grid place-items-center">
        {#if isSelected && !excludeMode && isBeingCompared}
          <CheckCircle color={getColor(row.index)} size="18px" />
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
