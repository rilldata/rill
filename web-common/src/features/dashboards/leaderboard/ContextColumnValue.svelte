<script lang="ts">
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import PercentageChange from "../../../components/data-types/PercentageChange.svelte";
  import { FormattedDataType } from "@rilldata/web-common/components/data-types";
  import { contextColumnWidth } from "./leaderboard-utils";

  export let formattedValue: string;
  export let showContext: LeaderboardContextColumn;
  $: neg = formattedValue[0] === "-";
  $: noData = formattedValue === "" || !formattedValue;
  $: customStyle = neg ? "text-red-500" : noData ? "opacity-50 italic" : "";
  $: {
    if (showContext === LeaderboardContextColumn.DELTA_ABSOLUTE) {
      console.log({ formattedValue, neg, noData });
    }
  }
  $: width = contextColumnWidth(showContext);
</script>

{#if showContext === LeaderboardContextColumn.DELTA_CHANGE || showContext === LeaderboardContextColumn.PERCENT}
  <div style:width>
    <PercentageChange value={formattedValue} />
  </div>
{:else if showContext === LeaderboardContextColumn.DELTA_ABSOLUTE}
  <div style:width>
    {#if noData}
      <span class="opacity-50 italic" style:font-size=".925em">no data</span>
    {:else}
      <FormattedDataType type="INTEGER" value={formattedValue} {customStyle} />
    {/if}
  </div>
{/if}
