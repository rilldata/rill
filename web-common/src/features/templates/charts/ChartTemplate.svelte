<script lang="ts">
  import Chart from "@rilldata/web-common/features/canvas-dashboards/Chart.svelte";
  import { ChartProperties } from "@rilldata/web-common/features/templates/types";
  import {
    V1ComponentSpecRendererProperties,
    V1ComponentVariable,
  } from "@rilldata/web-common/runtime-client";
  import { getSpecFromTemplateProperties } from "./utils";

  export let chartView: boolean;
  export let renderer: string;
  export let componentName: string;
  export let input: V1ComponentVariable[] | undefined;
  export let rendererProperties: V1ComponentSpecRendererProperties;

  $: chartProperties = rendererProperties as ChartProperties;
  $: vegaSpec = getSpecFromTemplateProperties(renderer, chartProperties);
</script>

<Chart {chartView} {vegaSpec} {input} {componentName} />
