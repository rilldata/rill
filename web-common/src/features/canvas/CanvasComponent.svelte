<script lang="ts" context="module">
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import ComponentRenderer from "@rilldata/web-common/features/canvas/components/ComponentRenderer.svelte";
  import {
    getComponentFilterProperties,
    isChartComponentType,
  } from "@rilldata/web-common/features/canvas/components/util";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
</script>

<script lang="ts">
  export let componentName: string;

  const {
    canvasEntity: {
      spec: { getComponentResourceFromName },
    },
  } = getCanvasStateManagers();

  $: component = getComponentResourceFromName(componentName);
  $: ({ renderer, rendererProperties } = $component ?? {});

  $: isChartType = isChartComponentType(renderer);

  $: title = rendererProperties?.title;
  $: description = rendererProperties?.description;
  $: componentFilters = getComponentFilterProperties(rendererProperties);
</script>

{#if !isChartType}
  <ComponentHeader {title} {description} filters={componentFilters} />
{/if}
{#if renderer && rendererProperties}
  <ComponentRenderer {renderer} {componentName} />
{/if}
