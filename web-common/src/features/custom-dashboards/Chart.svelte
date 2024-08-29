<script lang="ts">
  import { useVariableInputParams } from "@rilldata/web-common/features/custom-dashboards/variables-store";
  import {
    createQueryServiceResolveComponent,
    V1ComponentVariable,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getContext } from "svelte";
  import type { View, VisualizationSpec } from "svelte-vega";
  import VegaLiteRenderer from "../charts/render/VegaLiteRenderer.svelte";

  export let chartName: string;
  export let chartView: boolean;
  export let input: V1ComponentVariable[] | undefined;
  export let vegaSpec: VisualizationSpec | string | undefined;

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

  $: dashboardName = getContext("rill::custom-dashboard:name") as string;
  $: inputVariableParams = useVariableInputParams(dashboardName, input);

  $: chartDataQuery = createQueryServiceResolveComponent(
    $runtime.instanceId,
    chartName,
    { args: $inputVariableParams },
  );

  $: data = $chartDataQuery?.data?.data;
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
