<script lang="ts">
  import PercentageChange from "../../../components/data-types/PercentageChange.svelte";
  import { FormattedDataType } from "@rilldata/web-common/components/data-types";
  import { getStateManagers } from "../state-managers/state-managers";
  import type { LeaderboardItemData } from "./leaderboard-utils";
  import {
    formatMeasurePercentageDifference,
    formatProperFractionAsPercent,
  } from "../humanize-numbers";

  export let itemData: LeaderboardItemData;

  const {
    selectors: {
      contextColumn: {
        widthPx,
        isDeltaAbsolute,
        isDeltaPercent,
        isPercentOfTotal,
        isHidden,
      },
      numberFormat: { activeMeasureFormatter },
    },
  } = getStateManagers();

  $: negativeChange = itemData.deltaRel !== null && itemData.deltaAbs < 0;
  $: noChangeData = itemData.deltaRel === null;
</script>

{#if !$isHidden}
  <div style:width={$widthPx}>
    {#if $isPercentOfTotal}
      <PercentageChange
        value={formatProperFractionAsPercent(itemData.pctOfTotal)}
      />
    {:else if noChangeData}
      <span class="opacity-50 italic" style:font-size=".925em">no data</span>
    {:else if $isDeltaPercent}
      <PercentageChange
        value={formatMeasurePercentageDifference(itemData.deltaRel)}
      />
    {:else if $isDeltaAbsolute}
      <FormattedDataType
        type="INTEGER"
        value={$activeMeasureFormatter(itemData.deltaAbs)}
        customStyle={negativeChange ? "text-red-500" : ""}
      />
    {/if}
  </div>
{/if}
