<script lang="ts">
  import Chart from "@rilldata/web-common/features/custom-dashboards/Chart.svelte";
  import {
    V1ComponentSpecRendererProperties,
    V1ComponentSpecResolverProperties,
  } from "@rilldata/web-common/runtime-client";
  import { getSpecFromTemplateProperties } from "./utils";

  export let chartView: boolean;
  export let componentName: string;
  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let resolverProperties: V1ComponentSpecResolverProperties;

  $: vegaSpec = getSpecFromTemplateProperties(rendererProperties);
</script>

<Chart {chartView} {vegaSpec} chartName={componentName} {resolverProperties} />
