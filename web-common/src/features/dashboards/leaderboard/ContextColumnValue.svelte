<script lang="ts">
  import PercentageChange from "../../../components/data-types/PercentageChange.svelte";
  import { FormattedDataType } from "@rilldata/web-common/components/data-types";
  import { getStateManagers } from "../state-managers/state-managers";
  import type { LeaderboardItemData } from "./leaderboard-utils";
  import { formatProperFractionAsPercent } from "@rilldata/web-common/lib/number-formatting/proper-fraction-formatter";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { CONTEXT_COL_MAX_WIDTH } from "../state-managers/actions/context-columns";

  export let itemData: LeaderboardItemData;

  const {
    selectors: {
      contextColumn: {
        contextColumn,
        width,
        isDeltaAbsolute,
        isDeltaPercent,
        isPercentOfTotal,
        isHidden,
      },
      numberFormat: { activeMeasureFormatter },
    },
    actions: {
      contextCol: { observeContextColumnWidth },
    },
  } = getStateManagers();

  $: negativeChange = itemData.deltaAbs !== null && itemData.deltaAbs < 0;
  $: noChangeData = itemData.deltaRel === null;

  let element: HTMLElement;

  $: {
    // Re-observe the width when the context column changes,
    // but after a short delay to allow the DOM to update.
    if (element && $contextColumn) {
      setTimeout(() => {
        // the element may be gone by the time we get here,
        // if so, don't try to observe it
        if (!element) return;
        const this_width = element.getBoundingClientRect().width;

        // conditional, current store implementation
        if (this_width > $width && this_width < CONTEXT_COL_MAX_WIDTH) {
          observeContextColumnWidth(
            $contextColumn,
            element.getBoundingClientRect().width,
          );
        }

        // NOT conditional, current store implementation
        // observeContextColumnWidth(
        //   $contextColumn,
        //   element.getBoundingClientRect().width,
        // );
      }, 17);
    }
  }
</script>

{#if !$isHidden}
  <div style:width={$width + "px"} class="overflow-hidden">
    <div class="inline-block" bind:this={element}>
      {#if $isPercentOfTotal}
        <PercentageChange
          value={itemData.pctOfTotal
            ? formatProperFractionAsPercent(itemData.pctOfTotal)
            : null}
        />
      {:else if noChangeData}
        <span class="opacity-50 italic" style:font-size=".925em">no data</span>
      {:else if $isDeltaPercent}
        <PercentageChange
          value={itemData.deltaRel
            ? formatMeasurePercentageDifference(itemData.deltaRel)
            : null}
        />
      {:else if $isDeltaAbsolute}
        <FormattedDataType
          type="INTEGER"
          value={itemData.deltaAbs
            ? $activeMeasureFormatter(itemData.deltaAbs)
            : null}
          customStyle={negativeChange ? "text-red-500" : ""}
        />
      {/if}
    </div>
  </div>
{/if}
