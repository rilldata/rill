<script lang="ts">
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import PercentageChange from "../../../components/data-types/PercentageChange.svelte";
  import { FormattedDataType } from "@rilldata/web-common/components/data-types";
  import { contextColumnWidth } from "./leaderboard-utils";

  export let formattedValue: string;
  export let contextColumn: LeaderboardContextColumn;
  $: neg = formattedValue[0] === "-";
  $: noData = formattedValue === "" || !formattedValue;
  $: customStyle = neg ? "text-red-500" : noData ? "opacity-50 italic" : "";
  $: width = contextColumnWidth(contextColumn);
</script>

{#if contextColumn === LeaderboardContextColumn.DELTA_PERCENT || contextColumn === LeaderboardContextColumn.PERCENT}
  <div style:width="44px">
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
