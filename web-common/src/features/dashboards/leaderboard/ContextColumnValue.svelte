<script lang="ts">
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import PercentageChange from "../../../components/data-types/PercentageChange.svelte";
  import { FormattedDataType } from "@rilldata/web-common/components/data-types";
  import type { NumberParts } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { PERC_DIFF } from "@rilldata/web-common/components/data-types/type-utils";
  import { getStateManagers } from "../state-managers/state-managers";

  export let formattedValue: string | NumberParts | PERC_DIFF;
  // export let contextColumn: LeaderboardContextColumn;

  const stateManagers = getStateManagers();
  const { isAPercentColumn, isDeltaAbsolute, widthPx } =
    stateManagers.selectors.contextColumn;

  let neg: boolean;
  let noData: boolean;
  let customStyle: string;
  $: if (typeof formattedValue === "string") {
    neg = formattedValue[0] === "-";
    noData = formattedValue === "" || !formattedValue;
    customStyle = neg ? "text-red-500" : noData ? "opacity-50 italic" : "";
  }
  $: width = $widthPx;

  $: if (
    typeof formattedValue === "string" &&
    formattedValue !== PERC_DIFF.PREV_VALUE_NO_DATA
  ) {
    console.warn(
      `ContextColumnValue component expects a \`NumberParts | PERC_DIFF\`  received ${JSON.stringify(
        formattedValue
      )} instead.`
    );
  }
</script>

{#if $isAPercentColumn}
  <div style:width>
    <PercentageChange value={formattedValue} />
  </div>
{:else if $isDeltaAbsolute}
  <div style:width>
    {#if noData}
      <span class="opacity-50 italic" style:font-size=".925em">no data</span>
    {:else}
      <FormattedDataType type="INTEGER" value={formattedValue} {customStyle} />
    {/if}
  </div>
{/if}
