<script lang="ts">
  import Chart from "@rilldata/web-common/features/canvas/Chart.svelte";
  import { Image } from "./image";
  import { KPI } from "./kpi";
  import { Markdown } from "./markdown";
  import { Table } from "./table";

  import {
    createQueryServiceResolveComponent,
    type V1ComponentSpecRendererProperties,
    type V1ComponentSpecResolverProperties,
    type V1ComponentVariable,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let chartView: boolean;
  export let renderer: string;
  export let componentName: string;
  export let input: V1ComponentVariable[] | undefined;
  export let resolverProperties: V1ComponentSpecResolverProperties | undefined;

  $: componentQuery = createQueryServiceResolveComponent(
    $runtime.instanceId,
    componentName,
    { args: {} },
  );
  $: componentData = $componentQuery?.data;
  $: rendererProperties =
    componentData?.rendererProperties as V1ComponentSpecRendererProperties;
</script>

{#if rendererProperties}
  {#if renderer === "kpi"}
    <KPI {rendererProperties} />
  {:else if renderer === "table"}
    <Table {rendererProperties} />
  {:else if renderer === "markdown"}
    <Markdown {rendererProperties} />
  {:else if renderer === "image"}
    <Image {rendererProperties} />
  {:else if resolverProperties}
    <Chart {componentName} {chartView} {input} />
  {/if}
{/if}
