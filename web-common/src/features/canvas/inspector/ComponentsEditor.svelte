<script lang="ts">
  import {
    getHeaderForComponent,
    isCanvasComponentType,
  } from "@rilldata/web-common/features/canvas/components/util";
  import ParamMapper from "@rilldata/web-common/features/canvas/inspector/ParamMapper.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    ResourceKind,
    useResourceV2,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import SidebarWrapper from "@rilldata/web-common/features/visual-editing/SidebarWrapper.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let selectedComponentName: string;
  export let fileArtifact: FileArtifact;

  // TODO: Avoid resource query if possible
  $: resourceQuery = useResourceV2(
    $runtime.instanceId,
    selectedComponentName,
    ResourceKind.Component,
  );

  $: ({ data: componentResource } = $resourceQuery);

  $: ({ renderer, rendererProperties } =
    componentResource?.component?.spec ?? {});
</script>

<SidebarWrapper
  type="secondary"
  disableHorizontalPadding
  title={getHeaderForComponent(renderer)}
>
  {#if isCanvasComponentType(renderer) && rendererProperties}
    <ParamMapper
      {fileArtifact}
      componentType={renderer}
      paramValues={rendererProperties}
    />
  {:else}
    <div>
      Unknown Component {renderer}
    </div>
  {/if}
</SidebarWrapper>
