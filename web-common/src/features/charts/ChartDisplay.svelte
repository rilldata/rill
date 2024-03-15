<script lang="ts">
  import {
    ChartPromptStatus,
    chartPromptStore,
  } from "@rilldata/web-common/features/charts/prompt/chartPromptStatus";
  import VegaLiteRenderer from "@rilldata/web-common/features/charts/render/VegaLiteRenderer.svelte";
  import { useChart } from "@rilldata/web-common/features/charts/selectors";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createQuery } from "@tanstack/svelte-query";

  export let chartName: string;
  $: error = "";
  $: chart = useChart($runtime.instanceId, chartName);
  $: metricsQuery = $chart?.data?.chart?.spec?.resolverProperties;
  $: vegaSpec = $chart?.data?.chart?.spec?.vegaLiteSpec;
  $: data = {};

  async function fetchChartData(chartName: string) {
    // TODO: replace with prod API call
    const api_url = `http://localhost:9009/v1/instances/default/charts/${chartName}/data`;
    const response = await fetch(api_url);
    if (!response.ok) {
      error = `HTTP error! status: ${response.status}`;
      console.warn(response);
    }
    return response.json();
  }

  $: chartDataQuery = createQuery({
    queryKey: [`chart-data`, chartName, metricsQuery],
    queryFn: () => fetchChartData(chartName),
  });

  $: if (!$chartDataQuery.isFetching && $chartDataQuery?.data) {
    data = { table: $chartDataQuery?.data };
  }

  let parsedVegaSpec = undefined;
  $: try {
    parsedVegaSpec = vegaSpec ? JSON.parse(vegaSpec) : undefined;
    error = "";
  } catch (e) {
    error = e;
  }

  $: promptStatus =
    $chartPromptStore.charts[
      getFilePathFromNameAndType(chartName, EntityType.Chart)
    ];
</script>

<div class="m-2 w-1/2">
  {#if promptStatus}
    <p>
      Generating {promptStatus === ChartPromptStatus.GeneratingData
        ? "data"
        : "chart spec"} using AI
    </p>
  {:else if error}
    <p>{error}</p>
  {:else if !parsedVegaSpec}
    <p>Chart not available</p>
  {:else}
    <VegaLiteRenderer {data} spec={parsedVegaSpec} />
  {/if}
</div>
