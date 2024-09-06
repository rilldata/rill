<script lang="ts">
  import ChartsHeader from "@rilldata/web-common/features/charts/ChartsHeader.svelte";
  import ChartsEditor from "@rilldata/web-common/features/charts/editor/ChartsEditor.svelte";
  import ChartDataDisplay from "@rilldata/web-common/features/charts/prompt/ChartDataDisplay.svelte";
  import ChartStatusDisplay from "@rilldata/web-common/features/charts/prompt/ChartStatusDisplay.svelte";
  import CustomDashboardEmbed from "@rilldata/web-common/features/custom-dashboards/CustomDashboardEmbed.svelte";
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

  const dashboardName = getContext("rill::custom-dashboard:name") as string;

  let containerWidth: number;
  let containerHeight: number;
  let editorPercentage = 0.55;
  let tablePercentage = 0.45;

  $: ({ instanceId } = $runtime);

  $: ({ hasUnsavedChanges, path: filePath } = fileArtifact);
  $: chartName = getNameFromFile(filePath);

  $: editorWidth = editorPercentage * containerWidth;

  $: resourceQuery = useResource(instanceId, chartName, ResourceKind.Component);

  $: ({ data: componentResource, isFetching } = $resourceQuery);

  $: ({ resolverProperties, input } = componentResource?.component?.spec ?? {});
</script>

<WorkspaceContainer
  inspector={false}
  bind:width={containerWidth}
  bind:height={containerHeight}
>
  <ChartsHeader
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
      <ChartsEditor {filePath} />
    </div>
    <div class="size-full flex-col flex overflow-hidden">
      <ChartStatusDisplay {isFetching} {chartName}>
        <CustomDashboardEmbed
          {dashboardName}
          chartView
          gap={8}
          columns={10}
          items={[{ width: 10, height: 10, x: 0, y: 0, component: chartName }]}
        />
      </ChartStatusDisplay>

      <ChartDataDisplay
        {chartName}
        {tablePercentage}
        {containerHeight}
        {input}
        {resolverProperties}
      />
    </div>
  </div>
</WorkspaceContainer>
