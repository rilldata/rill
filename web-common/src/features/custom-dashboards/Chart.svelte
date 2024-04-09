<script lang="ts">
  import { useChart } from "@rilldata/web-common/features/charts/selectors";
  import DashVegaRenderer from "@rilldata/web-common/features/custom-dashboards/DashVegaRenderer.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { onDestroy, onMount } from "svelte";
  import type { VisualizationSpec } from "svelte-vega";
  import { createRuntimeServiceGetChartData } from "../../runtime-client/manual-clients";

  const observer = new ResizeObserver((entries) => {
    for (const entry of entries) {
      const { width, height } = entry.contentRect;
      clientHeight = height;
      clientWidth = width;
    }
  });
  const queryClient = useQueryClient();

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

  $: chartDataQuery = createRuntimeServiceGetChartData(
    queryClient,
    $runtime.instanceId,
    chartName,
    metricsQuery,
  );

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
