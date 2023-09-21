<script lang="ts">
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import PercentageChange from "../../../components/data-types/PercentageChange.svelte";
  import { FormattedDataType } from "@rilldata/web-common/components/data-types";
  import { contextColumnWidth } from "./leaderboard-utils";
  import type { NumberParts } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import type { PERC_DIFF } from "@rilldata/web-common/components/data-types/type-utils";

  export let formattedValue:
    | string
    | NumberParts
    | PERC_DIFF.PREV_VALUE_NO_DATA;
  export let contextColumn: LeaderboardContextColumn;

  let neg: boolean;
  let noData: boolean;
  let customStyle: string;
  $: if (typeof formattedValue === "string") {
    neg = formattedValue[0] === "-";
    noData = formattedValue === "" || !formattedValue;
    customStyle = neg ? "text-red-500" : noData ? "opacity-50 italic" : "";
  }
  $: width = contextColumnWidth(contextColumn);

  $: if (
    (contextColumn === LeaderboardContextColumn.DELTA_PERCENT ||
      contextColumn === LeaderboardContextColumn.PERCENT) &&
    typeof formattedValue === "string"
  ) {
    console.warn(
      "PercentageChange component expects a NumberParts object, not a string."
    );
  }
</script>

{#if contextColumn === LeaderboardContextColumn.DELTA_PERCENT || contextColumn === LeaderboardContextColumn.PERCENT}
  <div style:width>
    <PercentageChange value={formattedValue} />
  </div>
{:else if contextColumn === LeaderboardContextColumn.DELTA_ABSOLUTE}
  <div style:width>
    {#if noData}
      <span class="opacity-50 italic" style:font-size=".925em">no data</span>
    {:else}
      <FormattedDataType type="INTEGER" value={formattedValue} {customStyle} />
    {/if}
  </div>
{/if}
