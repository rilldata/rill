<script lang="ts">
  import { Image } from "./image";
  import { KPI } from "./kpi";
  import { Markdown } from "./markdown";
  import { Table } from "./table";

  import {
    createQueryServiceResolveComponent,
    type V1ComponentSpecRendererProperties,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let renderer: string;
  export let componentName: string;

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
  {/if}
{/if}
