<script lang="ts">
  import PercentageChange from "../../../components/data-types/PercentageChange.svelte";
  import { FormattedDataType } from "@rilldata/web-common/components/data-types";
  import { getStateManagers } from "../state-managers/state-managers";
  import type { LeaderboardItemData } from "./leaderboard-utils";
  import { formatProperFractionAsPercent } from "@rilldata/web-common/lib/number-formatting/proper-fraction-formatter";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";

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

  $: negativeChange = itemData.deltaAbs !== null && itemData.deltaAbs < 0;
  $: noChangeData = itemData.deltaRel === null;
</script>

{#if !$isHidden}
  <div style:width={$widthPx}>
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
{/if}
