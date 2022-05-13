<script lang="ts">
  import { getContext } from "svelte";
  import { min, max } from "d3-array";
  import { timeFormat } from "d3-time-format";
  import type { AppStore } from "$lib/application-state-stores/application-store";
  import ExploreChart from "$lib/components/viz/ExploreChart.svelte";
  import TopKSummary from "$lib/components/viz/TopKSummary.svelte";

  const store = getContext("rill:app:store") as AppStore;

  $: currentExploreConfiguration =
    $store && $store?.exploreConfigurations
      ? $store?.exploreConfigurations.find(
          (config) => config.id === $store?.activeAsset?.id
        )
      : undefined;

  $: preparedData = currentExploreConfiguration?.preview?.timeSeries.map(
    (d) => {
      d._ts = new Date(d._ts);
      d._ts.setHours(0, 0, 0, 0);
      return d;
    }
  );

  // generate some buttons.

  const fmt = timeFormat("%a %b %Y %I:%M:%S");

  $: latestDate = max(preparedData, (d) => d._ts);
  let thirtyDaysBefore;

  $: {
    thirtyDaysBefore = new Date(latestDate);
    thirtyDaysBefore.setDate(thirtyDaysBefore.getDate() - 31);
  }

  let ranges = { low: undefined, high: undefined };

  let hoveredDate;
</script>

<section>
  <div style:grid-area="main-header" class="p-5">
    <h2 class="text-lg">{currentExploreConfiguration?.name}</h2>
  </div>

  <div style:grid-area="timeseries-header" class="grid grid-flow-col">
    <button
      on:click={() => {
        //dataModelerService.dispatch('deleteExploreConfiguration', [{ id: currentExploreConfiguration.id }])
      }}>delete</button
    >

    <button
      on:click={() => {
        ranges = { low: undefined, high: undefined };
      }}
    >
      all time
    </button>

    <button
      on:click={() => {
        if (ranges.low) {
          const lowDate = new Date(ranges.low);
          lowDate.setDate(lowDate.getDate() - 28);
          const highDate = new Date(ranges.high);
          highDate.setDate(highDate.getDate() - 28);
          ranges = { low: lowDate, high: highDate };
        }
      }}
    >
      {"<"}
    </button>
    <button
      on:click={() => {
        if (ranges.low && ranges.high <= latestDate) {
          const lowDate = new Date(ranges.low);
          lowDate.setDate(lowDate.getDate() + 28);
          const highDate = new Date(ranges.high);
          highDate.setDate(highDate.getDate() + 28);

          ranges = { low: lowDate, high: highDate };
        }
      }}
    >
      {">"}
    </button>

    <button
      on:click={() => {
        ranges = { low: thirtyDaysBefore, high: latestDate };
      }}
    >
      last 30ish days
      <!-- {fmt(thirtyDaysBefore)} - {fmt(latestDate)} -->
    </button>
  </div>

  <div style:grid-area="leaderboard-header">leaderboard header</div>

  <div style:grid-area="timeseries-body">
    {#if currentExploreConfiguration?.preview}
      {#each currentExploreConfiguration.activeMetrics as metric, i}
        <ExploreChart
          data={preparedData}
          yAccessor={metric.aka}
          xMin={ranges.low}
          xMax={ranges.high}
          xAxis={i === 0}
          bind:hoveredDate
          zeroBound={metric.function === "count" ||
            metric.function === "approx_count_distinct"}
        />
      {/each}
    {/if}
  </div>

  <div style:grid-area="leaderboard-body" class="flex flex-wrap flex-col">
    {#each Object.keys(currentExploreConfiguration?.preview?.dimensionBoard) as k}
      {@const board = currentExploreConfiguration?.preview?.dimensionBoard[k]}
      <div style:width="300px">
        <h4>{board.column}</h4>
        <TopKSummary topK={board.topK.slice(0, 10)} totalRows={5000000} />
      </div>
    {/each}
  </div>
</section>

<style>
  section {
    display: grid;
    grid-column-gap: 2rem;
    /* grid-template-rows: auto auto auto; */
    grid-template-columns: [time-series] max-content [leaderboards] auto;
    grid-template-areas:
      "main-header main-header"
      "timeseries-header leaderboard-header"
      "timeseries-body leaderboard-body";
  }
</style>
