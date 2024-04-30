<script lang="ts">
  import { createRuntimeServiceGetChartData } from "@rilldata/web-common/runtime-client/manual-clients";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { VisualizationSpec } from "svelte-vega";
  import VegaLiteRenderer from "../charts/render/VegaLiteRenderer.svelte";
  import { V1ComponentSpecResolverProperties } from "@rilldata/web-common/runtime-client";

  export let chartName: string;
  export let chartView: boolean;
  export let vegaSpec: string;
  export let resolverProperties: V1ComponentSpecResolverProperties;

  let error: string | null = null;
  let parsedVegaSpec: VisualizationSpec | null = null;

  $: try {
    parsedVegaSpec = vegaSpec
      ? (JSON.parse(vegaSpec) as VisualizationSpec)
      : null;
    error = null;
  } catch (e: unknown) {
    error = JSON.stringify(e);
  }

  $: chartDataQuery = createRuntimeServiceGetChartData(
    queryClient,
    $runtime.instanceId,
    chartName,
    resolverProperties,
  );

  $: data = $chartDataQuery?.data;
</script>

{#if parsedVegaSpec}
  <VegaLiteRenderer
    customDashboard
    {error}
    {chartView}
    data={{ table: data }}
    spec={parsedVegaSpec}
  />
{/if}
