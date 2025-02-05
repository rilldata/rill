<script lang="ts">
  import {
    getComponentObj,
    getHeaderForComponent,
    isCanvasComponentType,
  } from "@rilldata/web-common/features/canvas/components/util";
  import VegaConfigInput from "@rilldata/web-common/features/canvas/inspector/VegaConfigInput.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import SidebarWrapper from "@rilldata/web-common/features/visual-editing/SidebarWrapper.svelte";
  import ComponentTabs from "./ComponentTabs.svelte";
  import FiltersMapper from "./filters/FiltersMapper.svelte";
  import ParamMapper from "./ParamMapper.svelte";

  export let selectedComponentIndex: number;
  export let fileArtifact: FileArtifact;

  const {
    canvasEntity: {
      spec: { getComponentFromIndex, getComponentNameFromIndex },
    },
  } = getCanvasStateManagers();
  let currentTab: string;

  $: componentSpec = getComponentFromIndex(selectedComponentIndex);
  $: componentName = getComponentNameFromIndex(selectedComponentIndex);

  $: ({ renderer, rendererProperties } = $componentSpec || {});

  $: componentType = isCanvasComponentType(renderer) ? renderer : null;
  $: path = ["items", selectedComponentIndex, "component", componentType || ""];

  $: component =
    componentType && rendererProperties
      ? getComponentObj(fileArtifact, path, componentType, rendererProperties)
      : null;
</script>

<SidebarWrapper
  type="secondary"
  disableHorizontalPadding
  title={getHeaderForComponent(componentType)}
>
  <svelte:fragment slot="header">
    {#if componentType}
      {#key componentType}
        <ComponentTabs {componentType} bind:currentTab />
      {/key}
    {/if}
  </svelte:fragment>

  {#if componentType && $componentName && component && rendererProperties}
    {#key $componentName}
      {#if currentTab === "options"}
        <ParamMapper
          {component}
          {componentType}
          paramValues={rendererProperties}
        />
      {:else if currentTab === "filters"}
        <FiltersMapper
          selectedComponentName={$componentName}
          {component}
          paramValues={rendererProperties}
        />
      {:else if currentTab === "config"}
        <VegaConfigInput {component} paramValues={rendererProperties} />
      {/if}
    {/key}
  {:else}
    <div>
      Unknown Component {renderer}
    </div>
  {/if}
</SidebarWrapper>
