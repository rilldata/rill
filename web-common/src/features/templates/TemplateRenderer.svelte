<script lang="ts">
  import ChartTemplate from "@rilldata/web-common/features/templates/charts/ChartTemplate.svelte";
  import KPITemplate from "@rilldata/web-common/features/templates/kpi/KPITemplate.svelte";
  import Markdown from "@rilldata/web-common/features/templates/markdown/Markdown.svelte";
  import Image from "@rilldata/web-common/features/templates/image/Image.svelte";

  import {
    V1ComponentSpecRendererProperties,
    V1ComponentSpecResolverProperties,
  } from "@rilldata/web-common/runtime-client";

  export let chartView: boolean;
  export let renderer: string;
  export let componentName: string;
  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let resolverProperties: V1ComponentSpecResolverProperties | undefined;
</script>

{#if renderer === "kpi"}
  <KPITemplate {rendererProperties} />
{:else if renderer === "markdown"}
  <Markdown {rendererProperties} />
{:else if renderer === "image"}
  <Image {rendererProperties} />
{:else if resolverProperties}
  <ChartTemplate
    {chartView}
    {renderer}
    {componentName}
    {rendererProperties}
    {resolverProperties}
  />
{/if}
