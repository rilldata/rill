<script lang="ts">
  import { useVariableInputParams } from "@rilldata/web-common/features/canvas-dashboards/variables-store";
  import {
    createQueryServiceResolveComponent,
    V1ComponentVariable,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getContext } from "svelte";
  import type { View, VisualizationSpec } from "svelte-vega";
  import VegaLiteRenderer from "../canvas-components/render/VegaLiteRenderer.svelte";

  export let componentName: string;
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

  $: dashboardName = getContext("rill::canvas-dashboard:name") as string;
  $: inputVariableParams = useVariableInputParams(dashboardName, input);

  $: componentDataQuery = createQueryServiceResolveComponent(
    $runtime.instanceId,
    componentName,
    { args: $inputVariableParams },
  );

  $: data = $componentDataQuery?.data?.data;
</script>

{#if parsedVegaSpec}
  <VegaLiteRenderer
    bind:viewVL
    canvasDashboard
    {error}
    {chartView}
    data={{ table: data }}
    spec={parsedVegaSpec}
  />
{/if}
