<script lang="ts">
  import ChartTemplate from "@rilldata/web-common/features/templates/charts/ChartTemplate.svelte";
  import Image from "@rilldata/web-common/features/templates/image/Image.svelte";
  import KPITemplate from "@rilldata/web-common/features/templates/kpi/KPITemplate.svelte";
  import Markdown from "@rilldata/web-common/features/templates/markdown/Markdown.svelte";
  import Select from "@rilldata/web-common/features/templates/select/Select.svelte";
  import Switch from "@rilldata/web-common/features/templates/switch/Switch.svelte";
  import TableTemplate from "@rilldata/web-common/features/templates/table/TableTemplate.svelte";

  import {
    V1ComponentSpecRendererProperties,
    V1ComponentSpecResolverProperties,
    V1ComponentVariable,
  } from "@rilldata/web-common/runtime-client";

  export let chartView: boolean;
  export let renderer: string;
  export let componentName: string;
  export let input: V1ComponentVariable[] | undefined;
  export let output: V1ComponentVariable | undefined;
  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let resolverProperties: V1ComponentSpecResolverProperties | undefined;
</script>

{#if renderer === "kpi"}
  <KPITemplate {rendererProperties} />
{:else if renderer === "table"}
  <TableTemplate {rendererProperties} />
{:else if renderer === "markdown"}
  <Markdown {rendererProperties} />
{:else if renderer === "image"}
  <Image {rendererProperties} />
{:else if renderer === "select"}
  <Select
    {componentName}
    {input}
    {output}
    {resolverProperties}
    {rendererProperties}
  />
{:else if renderer === "switch"}
  <Switch {output} {rendererProperties} />
{:else if resolverProperties}
  <ChartTemplate
    {chartView}
    {input}
    {renderer}
    {componentName}
    {rendererProperties}
    {resolverProperties}
  />
{/if}
