<script lang="ts">
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import ComponentStatusDisplay from "@rilldata/web-common/features/canvas-components/ComponentStatusDisplay.svelte";
  import ComponentsHeader from "@rilldata/web-common/features/canvas-components/ComponentsHeader.svelte";
  import ComponentsEditor from "@rilldata/web-common/features/canvas-components/editor/ComponentsEditor.svelte";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas-dashboards/CanvasDashboardEmbed.svelte";
  import { useVariableInputParams } from "@rilldata/web-common/features/canvas-dashboards/variables-store";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { WorkspaceContainer } from "@rilldata/web-common/layout/workspace";
  import {
    createQueryServiceResolveComponent,
    V1MetricsViewRowsResponseDataItem,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getContext } from "svelte";

  export let fileArtifact: FileArtifact;

  const dashboardName = getContext("rill::canvas-dashboard:name") as string;

  let containerWidth: number;
  let containerHeight: number;
  let editorPercentage = 0.55;
  let tablePercentage = 0.45;

  $: ({ instanceId } = $runtime);

  $: ({ hasUnsavedChanges, path: filePath } = fileArtifact);
  $: componentName = getNameFromFile(filePath);

  $: editorWidth = editorPercentage * containerWidth;
  $: tableHeight = tablePercentage * containerHeight;

  $: resourceQuery = useResource(
    instanceId,
    componentName,
    ResourceKind.Component,
  );

  $: ({ data: componentResource } = $resourceQuery);

  $: ({ resolverProperties, input } = componentResource?.component?.spec ?? {});

  $: inputVariableParams = useVariableInputParams(dashboardName, input);

  $: componentDataQuery = resolverProperties
    ? createQueryServiceResolveComponent(instanceId, componentName, {
        args: $inputVariableParams,
      })
    : null;

  let isFetching = false;
  let componentData: V1MetricsViewRowsResponseDataItem[] | undefined =
    undefined;

  $: if (componentDataQuery) {
    isFetching = $componentDataQuery?.isFetching ?? false;
    componentData = $componentDataQuery?.data?.data;
  }
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
    <div
      style:width="{editorPercentage * 100}%"
      class="relative flex-none border-r"
    >
      <Resizer
        direction="EW"
        side="right"
        dimension={editorWidth}
        min={300}
        max={0.65 * containerWidth}
        onUpdate={(width) => (editorPercentage = width / containerWidth)}
      />
      <ComponentsEditor {filePath} />
    </div>
    <div class="size-full flex-col flex overflow-hidden">
      <ComponentStatusDisplay {isFetching} {componentName}>
        <CanvasDashboardEmbed
          {dashboardName}
          chartView
          gap={8}
          columns={10}
          items={[
            { width: 10, height: 10, x: 0, y: 0, component: componentName },
          ]}
        />
      </ComponentStatusDisplay>

      {#if componentDataQuery}
        <div
          class="size-full h-48 bg-gray-100 border-t relative flex-none flex-shrink-0"
          style:height="{tablePercentage * 100}%"
        >
          <Resizer
            direction="NS"
            dimension={tableHeight}
            min={100}
            max={0.65 * containerHeight}
            onUpdate={(height) => (tablePercentage = height / containerHeight)}
          />

          {#if isFetching}
            <div
              class="flex flex-col gap-y-2 size-full justify-center items-center"
            >
              <Spinner size="2em" status={EntityStatus.Running} />
              <div>Loading component data</div>
            </div>
          {:else if componentData}
            <PreviewTable
              rows={componentData}
              name={componentName}
              columnNames={Object.keys(componentData[0]).map((key) => ({
                type: "VARCHAR",
                name: key,
              }))}
            />
          {:else}
            <p class="text-lg size-full grid place-content-center">
              Update YAML to view component data
            </p>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</WorkspaceContainer>
