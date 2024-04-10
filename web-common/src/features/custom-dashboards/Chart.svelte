<script lang="ts">
  import DashVegaRenderer from "@rilldata/web-common/features/custom-dashboards/DashVegaRenderer.svelte";
  import { useChart } from "@rilldata/web-common/features/charts/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onDestroy, onMount } from "svelte";
  import type { VisualizationSpec } from "svelte-vega";
  import { createQuery } from "@tanstack/svelte-query";

  const observer = new ResizeObserver((entries) => {
    for (const entry of entries) {
      const { width, height } = entry.contentRect;
      clientHeight = height;
      clientWidth = width;
    }
  });

  export let chartName: string;

  let clientHeight: number;
  let clientWidth: number;
  let container: HTMLDivElement;
  let error: unknown = "";
  let parsedVegaSpec: VisualizationSpec | undefined = undefined;

  $: chart = useChart($runtime.instanceId, chartName);

  $: vegaSpec = $chart?.data?.chart?.spec?.vegaLiteSpec;

  $: try {
    parsedVegaSpec = vegaSpec
      ? (JSON.parse(vegaSpec) as VisualizationSpec)
      : undefined;
    error = "";
  } catch (e: unknown) {
    error = e;
  }

  $: metricsQuery = $chart?.data?.chart?.spec?.resolverProperties;

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

  $: data = $chartDataQuery?.data;

  onMount(() => {
    observer.observe(container);
  });

  onDestroy(() => {
    observer.disconnect();
  });
</script>

<div
  class="h-full w-full overflow-hidden pointer-events-none"
  bind:this={container}
>
  {#if error}
    <p>{error}</p>
  {:else if !parsedVegaSpec}
    <p>Chart not available</p>
  {:else}
    <DashVegaRenderer
      data={{ table: data }}
      spec={parsedVegaSpec}
      height={clientHeight - 31}
      width={clientWidth}
    />
  {/if}
</div>
