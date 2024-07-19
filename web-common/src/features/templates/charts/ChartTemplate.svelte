<script lang="ts">
  import Chart from "@rilldata/web-common/features/custom-dashboards/Chart.svelte";
  import { ChartProperties } from "@rilldata/web-common/features/templates/types";
  import {
    V1ComponentSpecRendererProperties,
    V1ComponentSpecResolverProperties,
  } from "@rilldata/web-common/runtime-client";
  import { getSpecFromTemplateProperties } from "./utils";

  export let chartView: boolean;
  export let renderer: string;
  export let componentName: string;
  export let input: string[] | undefined;
  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let resolverProperties: V1ComponentSpecResolverProperties;

  $: chartProperties = rendererProperties as ChartProperties;
  $: vegaSpec = getSpecFromTemplateProperties(renderer, chartProperties);
</script>

<Chart
  {chartView}
  {vegaSpec}
  {input}
  chartName={componentName}
  {resolverProperties}
/>
