<script lang="ts">
  import { useChart } from "@rilldata/web-common/features/charts/selectors";
  import { createRuntimeServiceGetChartData } from "@rilldata/web-common/runtime-client/manual-clients";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { VisualizationSpec } from "svelte-vega";
  import VegaLiteRenderer from "../charts/render/VegaLiteRenderer.svelte";

  export let chartName: string;

  let error: string | null = null;
  let parsedVegaSpec: VisualizationSpec | null = null;

  $: chart = useChart($runtime.instanceId, chartName);

  $: vegaSpec = $chart?.data?.chart?.spec?.vegaLiteSpec;

  $: try {
    parsedVegaSpec = vegaSpec
      ? (JSON.parse(vegaSpec) as VisualizationSpec)
      : null;
    error = null;
  } catch (e: unknown) {
    error = JSON.stringify(e);
  }

  $: metricsQuery = $chart?.data?.chart?.spec?.resolverProperties;

  $: chartDataQuery = createRuntimeServiceGetChartData(
    queryClient,
    $runtime.instanceId,
    chartName,
    metricsQuery,
  );

  $: data = $chartDataQuery?.data;
</script>

{#if parsedVegaSpec}
  <VegaLiteRenderer
    dashboard
    data={{ table: data }}
    spec={parsedVegaSpec}
    {error}
  />
{/if}
