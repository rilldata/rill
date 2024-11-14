<script lang="ts">
  import ComponentDataDisplay from "@rilldata/web-common/features/canvas-components/ComponentDataDisplay.svelte";
  import ComponentStatusDisplay from "@rilldata/web-common/features/canvas-components/ComponentStatusDisplay.svelte";
  import ComponentsHeader from "@rilldata/web-common/features/canvas-components/ComponentsHeader.svelte";
  import ComponentsEditor from "@rilldata/web-common/features/canvas-components/editor/ComponentsEditor.svelte";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { WorkspaceContainer } from "@rilldata/web-common/layout/workspace";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getContext } from "svelte";

  export let fileArtifact: FileArtifact;

  const canvasName = getContext("rill::canvas:name") as string;

  let containerWidth: number;
  let containerHeight: number;
  let editorPercentage = 0.55;
  let tablePercentage = 0.45;

  $: ({ instanceId } = $runtime);

  $: ({ hasUnsavedChanges, path: filePath } = fileArtifact);
  $: componentName = getNameFromFile(filePath);

  $: editorWidth = editorPercentage * containerWidth;

  $: resourceQuery = useResource(
    instanceId,
    componentName,
    ResourceKind.Component,
  );

  $: ({ data: componentResource, isFetching } = $resourceQuery);

  $: ({ resolverProperties, input } = componentResource?.component?.spec ?? {});

  $: tableHeight = tablePercentage * containerHeight;
</script>

<WorkspaceContainer
  inspector={false}
  bind:width={containerWidth}
  bind:height={containerHeight}
>
  <ComponentsHeader
    slot="header"
    {filePath}
    hasUnsavedChanges={$hasUnsavedChanges}
  />
  <div slot="body" class="flex size-full">
    <section style:width="{editorPercentage * 100}%">
      <ComponentsEditor {filePath} />
    </section>

    <Resizer
      absolute={false}
      direction="EW"
      side="right"
      dimension={editorWidth}
      min={300}
      max={0.65 * containerWidth}
      onUpdate={(width) => (editorPercentage = width / containerWidth)}
    />

    <section class="size-full flex-col flex overflow-hidden">
      <div
        class="size-full overflow-hidden border border-gray-300 rounded-[2px]"
      >
        <ComponentStatusDisplay {isFetching} {componentName}>
          <CanvasDashboardEmbed
            {canvasName}
            chartView
            gap={8}
            columns={10}
            items={[
              { width: 10, height: 10, x: 0, y: 0, component: componentName },
            ]}
          />
        </ComponentStatusDisplay>
      </div>

      <Resizer
        absolute={false}
        direction="NS"
        dimension={tableHeight}
        min={100}
        max={0.65 * containerHeight}
        onUpdate={(height) => (tablePercentage = height / containerHeight)}
      />

      <ComponentDataDisplay
        {componentName}
        {tablePercentage}
        {input}
        {resolverProperties}
      />
    </section>
  </div>
</WorkspaceContainer>
