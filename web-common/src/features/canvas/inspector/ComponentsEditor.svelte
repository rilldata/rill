<script lang="ts">
  import {
    getComponentObj,
    getHeaderForComponent,
    isCanvasComponentType,
  } from "@rilldata/web-common/features/canvas/components/util";
  import VegaConfigInput from "@rilldata/web-common/features/canvas/inspector/VegaConfigInput.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    ResourceKind,
    useResourceV2,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import SidebarWrapper from "@rilldata/web-common/features/visual-editing/SidebarWrapper.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ComponentTabs from "./ComponentTabs.svelte";
  import FiltersMapper from "./FiltersMapper.svelte";
  import ParamMapper from "./ParamMapper.svelte";

  export let selectedComponentName: string;
  export let fileArtifact: FileArtifact;

  const ctx = getCanvasStateManagers();
  let currentTab: string;

  // TODO: Avoid resource query if possible
  $: resourceQuery = useResourceV2(
    $runtime.instanceId,
    selectedComponentName,
    ResourceKind.Component,
  );

  $: ({ data: componentResource } = $resourceQuery);

  $: ({ renderer, rendererProperties } =
    componentResource?.component?.spec ?? {});

  $: componentType = isCanvasComponentType(renderer) ? renderer : null;

  $: selectedIndexStore = ctx.canvasEntity?.selectedComponentIndex;
  $: selectedComponentIndex = $selectedIndexStore ?? 0;
  $: path = ["items", selectedComponentIndex, "component", componentType || ""];

  $: component =
    componentType && rendererProperties
      ? getComponentObj(fileArtifact, path, componentType, rendererProperties)
      : null;
</script>

<SidebarWrapper
  type="secondary"
  disableHorizontalPadding
  title={getHeaderForComponent(renderer)}
>
  <svelte:fragment slot="header">
    {#if componentType}
      {#key componentType}
        <ComponentTabs {componentType} bind:currentTab />
      {/key}
    {/if}
  </svelte:fragment>

  {#if componentType && component && rendererProperties}
    {#key selectedComponentIndex}
      {#if currentTab === "options"}
        <ParamMapper
          {component}
          {componentType}
          paramValues={rendererProperties}
        />
      {:else if currentTab === "filters"}
        <FiltersMapper {component} paramValues={rendererProperties} />
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
