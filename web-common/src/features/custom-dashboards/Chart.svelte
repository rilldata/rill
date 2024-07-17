<script lang="ts">
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { V1ComponentSpecResolverProperties } from "@rilldata/web-common/runtime-client";
  import { createRuntimeServiceGetChartData } from "@rilldata/web-common/runtime-client/manual-clients";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { View, VisualizationSpec } from "svelte-vega";
  import VegaLiteRenderer from "../charts/render/VegaLiteRenderer.svelte";

  export let chartName: string;
  export let chartView: boolean;
  export let vegaSpec: VisualizationSpec | string | undefined;
  export let resolverProperties: V1ComponentSpecResolverProperties;

  let viewVL: View;
  let error: string | null = null;
  let parsedVegaSpec: VisualizationSpec | null = null;

  $: try {
    if (typeof vegaSpec === "string") {
      parsedVegaSpec = JSON.parse(vegaSpec) as VisualizationSpec;
    } else {
      parsedVegaSpec = vegaSpec ?? null;
    }
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
    bind:viewVL
    customDashboard
    {error}
    {chartView}
    data={{ table: data }}
    spec={parsedVegaSpec}
  />
{/if}
