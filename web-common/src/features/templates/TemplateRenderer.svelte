<script lang="ts">
  import Chart from "@rilldata/web-common/features/canvas/Chart.svelte";
  import { useVariableInputParams } from "@rilldata/web-common/features/canvas/variables-store";
  import Image from "@rilldata/web-common/features/templates/image/Image.svelte";
  import KPITemplate from "@rilldata/web-common/features/templates/kpi/KPITemplate.svelte";
  import Markdown from "@rilldata/web-common/features/templates/markdown/Markdown.svelte";
  import Select from "@rilldata/web-common/features/templates/select/Select.svelte";
  import Switch from "@rilldata/web-common/features/templates/switch/Switch.svelte";
  import TableTemplate from "@rilldata/web-common/features/templates/table/TableTemplate.svelte";

  import {
    createQueryServiceResolveComponent,
    type V1ComponentSpecRendererProperties,
    type V1ComponentSpecResolverProperties,
    type V1ComponentVariable,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getContext } from "svelte";
  import { RendererType } from "./types";

  export let chartView: boolean;
  export let renderer: string;
  export let componentName: string;
  export let input: V1ComponentVariable[] | undefined;
  export let output: V1ComponentVariable | undefined;
  export let resolverProperties: V1ComponentSpecResolverProperties | undefined;

  const canvasName = getContext("rill::canvas:name") as string;

  $: inputVariableParams = useVariableInputParams(canvasName, input);
  $: componentQuery = createQueryServiceResolveComponent(
    $runtime.instanceId,
    componentName,
    { args: $inputVariableParams },
  );
  $: componentData = $componentQuery?.data;
  $: rendererProperties =
    componentData?.rendererProperties as V1ComponentSpecRendererProperties;
  $: data = componentData?.data;
</script>

{#if rendererProperties}
  {#if renderer === RendererType.KPI}
    <KPITemplate {rendererProperties} />
  {:else if renderer === RendererType.Table}
    <TableTemplate {rendererProperties} />
  {:else if renderer === RendererType.Markdown}
    <Markdown {rendererProperties} />
  {:else if renderer === RendererType.Image}
    <Image {rendererProperties} />
  {:else if renderer === RendererType.Select}
    <Select {data} {componentName} {output} {rendererProperties} />
  {:else if renderer === RendererType.Switch}
    <Switch {output} {rendererProperties} />
  {:else if resolverProperties}
    <Chart {componentName} {chartView} {input} />
  {/if}
{/if}
