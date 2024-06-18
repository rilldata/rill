<script lang="ts">
  import ChartTemplate from "@rilldata/web-common/features/templates/charts/ChartTemplate.svelte";
  import KPITemplate from "@rilldata/web-common/features/templates/kpi/KPITemplate.svelte";

  import {
    V1ComponentSpecRendererProperties,
    V1ComponentSpecResolverProperties,
  } from "@rilldata/web-common/runtime-client";

  export let chartView: boolean;
  export let renderer: string;
  export let componentName: string;
  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let resolverProperties: V1ComponentSpecResolverProperties;
</script>

{#if renderer === "kpi"}
  <KPITemplate {rendererProperties} />
{:else}
  <ChartTemplate
    {chartView}
    {componentName}
    {rendererProperties}
    {resolverProperties}
  />
{/if}
