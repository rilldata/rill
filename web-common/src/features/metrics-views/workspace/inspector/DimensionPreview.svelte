<script>
  import LeaderboardListItem from "@rilldata/web-common/features/dashboards/leaderboard/LeaderboardListItem.svelte";
  import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors";
  import { createQueryServiceMetricsViewToplist } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let metricsDefName;
  export let dimensionName;

  $: metaQuery = useMetaQuery($runtime.instanceId, metricsDefName);
  $: leaderboard = createQueryServiceMetricsViewToplist(
    $runtime.instanceId,
    metricsDefName,
    {
      dimensionName: dimensionName,
      measureNames: $metaQuery?.data?.measures?.map((m) => m.name),
      limit: `3`,
      offset: "0",
      sort: [],
      // sort: [
      //   {
      //     name: $metaQuery?.data?.measures?.[0]?.name,
      //     ascending: false,
      //   },
      // ],
    },
    {
      query: { enabled: $metaQuery?.data?.measures?.length > 0 },
    }
  );
  $: console.log(dimensionName, $leaderboard?.data);
</script>

<div>
  <div class="px-4">{dimensionName}</div>
  {#if $leaderboard?.data?.data}
    {#each $leaderboard?.data?.data as item}
      <LeaderboardListItem showIcon={false}>
        <div class="pl-4" slot="title">{item[dimensionName]}</div>
        <div class="pr-4" slot="right">{item.measure_0}</div>
      </LeaderboardListItem>
    {/each}
  {/if}
</div>
