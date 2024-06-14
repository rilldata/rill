<script lang="ts">
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import ChartsHeader from "@rilldata/web-common/features/charts/ChartsHeader.svelte";
  import ChartsEditor from "@rilldata/web-common/features/charts/editor/ChartsEditor.svelte";
  import ChartPromptStatusDisplay from "@rilldata/web-common/features/charts/prompt/ChartPromptStatusDisplay.svelte";
  import CustomDashboardEmbed from "@rilldata/web-common/features/custom-dashboards/CustomDashboardEmbed.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { WorkspaceContainer } from "@rilldata/web-common/layout/workspace";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { createRuntimeServiceGetChartData } from "@rilldata/web-common/runtime-client/manual-clients";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";

  export let fileArtifact: FileArtifact;

  let containerWidth: number;
  let containerHeight: number;
  let editorPercentage = 0.55;
  let tablePercentage = 0.45;

  $: ({ instanceId } = $runtime);

  $: ({ hasUnsavedChanges, path: filePath } = fileArtifact);
  $: chartName = getNameFromFile(filePath);

  $: editorWidth = editorPercentage * containerWidth;
  $: tableHeight = tablePercentage * containerHeight;

  $: resourceQuery = useResource(instanceId, chartName, ResourceKind.Component);

  $: ({ data: componentResource } = $resourceQuery);

  $: ({ resolverProperties } = componentResource?.component?.spec ?? {});

  $: chartDataQuery = createRuntimeServiceGetChartData(
    queryClient,
    instanceId,
    chartName,
    resolverProperties,
  );

  $: ({ isFetching: chartDataFetching, data: chartData } = $chartDataQuery);
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
      <ChartPromptStatusDisplay {chartName}>
        <CustomDashboardEmbed
          chartView
          gap={8}
          columns={10}
          items={[{ width: 10, height: 10, x: 0, y: 0, component: chartName }]}
        />
      </ChartPromptStatusDisplay>

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

        {#if chartDataFetching}
          <div
            class="flex flex-col gap-y-2 size-full justify-center items-center"
          >
            <Spinner size="2em" status={EntityStatus.Running} />
            <div>Loading chart data</div>
          </div>
        {:else if chartData}
          <PreviewTable
            rows={chartData}
            name={chartName}
            columnNames={Object.keys(chartData[0]).map((key) => ({
              type: "VARCHAR",
              name: key,
            }))}
          />
        {:else}
          <p class="text-lg size-full grid place-content-center">
            Update YAML to view chart data
          </p>
        {/if}
      </div>
    </div>
  </div>
</WorkspaceContainer>
