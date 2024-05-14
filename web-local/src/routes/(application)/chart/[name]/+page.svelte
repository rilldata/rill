<script lang="ts">
  import { page } from "$app/stores";
  import ChartsHeader from "@rilldata/web-common/features/charts/ChartsHeader.svelte";
  import ChartsEditor from "@rilldata/web-common/features/charts/editor/ChartsEditor.svelte";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import {
    EntityStatus,
    EntityType,
  } from "@rilldata/web-common/features/entity-management/types";
  import { WorkspaceContainer } from "@rilldata/web-common/layout/workspace";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { CATALOG_ENTRY_NOT_FOUND } from "@rilldata/web-local/lib/errors/messages";
  import { error } from "@sveltejs/kit";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import ChartPromptStatusDisplay from "@rilldata/web-common/features/charts/prompt/ChartPromptStatusDisplay.svelte";
  import CustomDashboardEmbed from "@rilldata/web-common/features/custom-dashboards/CustomDashboardEmbed.svelte";
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { createRuntimeServiceGetChartData } from "@rilldata/web-common/runtime-client/manual-clients";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";

  export let data: { fileArtifact?: FileArtifact } = {};

  let containerWidth: number;
  let containerHeight: number;
  let editorPercentage = 0.55;
  let tablePercentage = 0.45;
  let filePath: string;
  let chartName: string;

  $: workspace = workspaces.get(filePath);
  $: autoSave = workspace.editor.autoSave;

  if (data.fileArtifact) {
    filePath = data.fileArtifact.path;
    chartName = data.fileArtifact.getEntityName();
  } else {
    // needed for backwards compatibility for now
    chartName = $page.params.name;
    filePath = getFileAPIPathFromNameAndType(chartName, EntityType.Chart);
  }

  $: fileQuery = createRuntimeServiceGetFile(
    instanceId,
    {
      path: filePath,
    },
    {
      query: {
        onError: (err) => {
          if (err.response?.data?.message.includes(CATALOG_ENTRY_NOT_FOUND)) {
            throw error(404, "Dashboard not found");
          }

          throw error(err.response?.status || 500, err.message);
        },
        refetchOnWindowFocus: false,
      },
    },
  );

  $: ({ instanceId } = $runtime);

  $: yaml = $fileQuery.data?.blob || "";
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

<svelte:head>
  <title>Rill Developer | {chartName}</title>
</svelte:head>

{#if $fileQuery.data && yaml !== undefined}
  <WorkspaceContainer
    inspector={false}
    bind:width={containerWidth}
    bind:height={containerHeight}
  >
    <ChartsHeader slot="header" {filePath} />
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
        <ChartsEditor {filePath} bind:autoSave={$autoSave} />
      </div>
      <div class="size-full flex-col flex overflow-hidden">
        <ChartPromptStatusDisplay {chartName}>
          <CustomDashboardEmbed
            chartView
            gap={8}
            columns={10}
            items={[
              { width: 10, height: 10, x: 0, y: 0, component: chartName },
            ]}
          />
        </ChartPromptStatusDisplay>

        <div
          class="size-full h-48 bg-gray-100 border-t relative flex-none flex-shrink-0 grid place-content-center"
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
            <div class="flex flex-col gap-y-2 items-center">
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
            <p class="text-lg">Update YAML to view chart data</p>
          {/if}
        </div>
      </div>
    </div>
  </WorkspaceContainer>
{/if}
