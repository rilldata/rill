<script lang="ts">
  import {
    getHeaderForComponent,
    isCanvasComponentType,
  } from "@rilldata/web-common/features/canvas/components/util";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import SidebarWrapper from "@rilldata/web-common/features/visual-editing/SidebarWrapper.svelte";
  import VegaConfigInput from "./chart/VegaConfigInput.svelte";
  import ComponentTabs from "./ComponentTabs.svelte";
  import FiltersMapper from "./filters/FiltersMapper.svelte";
  import ParamMapper from "./ParamMapper.svelte";
  import { hasComponentFilters } from "./util";
  import type { BaseCanvasComponent } from "../components/BaseCanvasComponent";

  export let canvasName: string;
  export let component: BaseCanvasComponent;

  let currentTab: string;

  $: ({ specStore, type } = component);

  $: rendererProperties = $specStore;

  $: componentType = isCanvasComponentType(type) ? type : null;
</script>

<SidebarWrapper
  type="secondary"
  disableHorizontalPadding
  title={getHeaderForComponent(componentType)}
>
  {component.pathInYAML}
  <svelte:fragment slot="header">
    {#if componentType}
      {#key componentType}
        <ComponentTabs
          hasFilters={hasComponentFilters(component)}
          {componentType}
          bind:currentTab
        />
      {/key}
    {/if}
  </svelte:fragment>

  {#if componentType && component && rendererProperties}
    {#key component}
      {#if currentTab === "options"}
        <ParamMapper
          {canvasName}
          {component}
          paramValues={rendererProperties}
        />
      {:else if currentTab === "filters"}
        <FiltersMapper
          {canvasName}
          {component}
          paramValues={rendererProperties}
        />
      {:else if currentTab === "config"}
        <VegaConfigInput {component} paramValues={rendererProperties} />
      {/if}
    {/key}
  {:else if !type}
    <div class="inspector-center">
      <Spinner status={EntityStatus.Running} size="16px" />
    </div>
  {:else}
    <div class="inspector-center">
      Unknown Component {type}
    </div>
  {/if}
</SidebarWrapper>

<style lang="postcss">
  .inspector-center {
    @apply flex items-center justify-center h-full w-full;
  }
</style>
