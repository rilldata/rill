<script lang="ts">
  import { ExternalLink } from "lucide-svelte";
  import {
    getHeaderForComponent,
    isCanvasComponentType,
  } from "@rilldata/web-common/features/canvas/components/util";
  import { ComponentRefComponent } from "@rilldata/web-common/features/canvas/components/component-ref";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import SidebarWrapper from "@rilldata/web-common/features/visual-editing/SidebarWrapper.svelte";
  import { getFileHref } from "@rilldata/web-common/layout/navigation/editor-routing";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { BaseCanvasComponent } from "../components/BaseCanvasComponent";
  import VegaConfigInput from "./chart/VegaConfigInput.svelte";
  import ComponentTabs from "./ComponentTabs.svelte";
  import FiltersMapper from "./filters/FiltersMapper.svelte";
  import ParamMapper from "./ParamMapper.svelte";
  import { hasComponentFilters } from "./util";

  export let component: BaseCanvasComponent;

  let currentTab: string;

  $: ({ specStore, type, resource } = component);

  $: rendererProperties = $specStore;

  $: componentType = isCanvasComponentType(type) ? type : null;

  // For items referencing an externally defined component: link to the component's file.
  $: isRef = component instanceof ComponentRefComponent;
  $: refDisplayName =
    $resource?.component?.state?.validSpec?.displayName ??
    $resource?.meta?.name?.name ??
    "";
  $: refFilePath = $resource?.meta?.filePaths?.[0];
</script>

<SidebarWrapper
  type="secondary"
  disableHorizontalPadding
  title={isRef ? refDisplayName : getHeaderForComponent(componentType)}
>
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

  {#if isRef}
    <div class="flex flex-col gap-y-1 px-5 py-3 border-b">
      {#if refFilePath}
        <a
          class="flex items-center gap-x-1 text-xs font-medium text-primary-600 hover:text-primary-700 w-fit"
          href={getFileHref(refFilePath)}
        >
          {m.canvas_edit_component()}
          <ExternalLink size="12px" />
        </a>
      {/if}
      <span class="text-[11px] text-fg-secondary">
        {m.canvas_component_ref_edit_note()}
      </span>
    </div>
  {/if}

  {#if componentType && component && rendererProperties && currentTab}
    {#if currentTab === "options"}
      {#key component}
        <ParamMapper {component} />
      {/key}
    {:else if currentTab === "filters"}
      <FiltersMapper {component} />
    {:else if currentTab === "config"}
      <VegaConfigInput {component} />
    {/if}
  {:else if !type}
    <div class="inspector-center">
      <Spinner status={EntityStatus.Running} size="16px" />
    </div>
  {:else}
    <div class="inspector-center">
      {m.canvas_inspector_unknown_component({ type })}
    </div>
  {/if}
</SidebarWrapper>

<style lang="postcss">
  .inspector-center {
    @apply flex items-center justify-center h-full w-full;
  }
</style>
