<script lang="ts">
  import { page } from "$app/stores";
  import ChartsHeader from "@rilldata/web-common/features/charts/ChartsHeader.svelte";
  import ChartsEditor from "@rilldata/web-common/features/charts/editor/ChartsEditor.svelte";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { WorkspaceContainer } from "@rilldata/web-common/layout/workspace";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { CATALOG_ENTRY_NOT_FOUND } from "@rilldata/web-local/lib/errors/messages";
  import { error } from "@sveltejs/kit";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import ChartPromptStatusDisplay from "@rilldata/web-common/features/charts/prompt/ChartPromptStatusDisplay.svelte";
  import CustomDashboardEmbed from "@rilldata/web-common/features/custom-dashboards/CustomDashboardEmbed.svelte";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { createRuntimeServiceGetChartData } from "@rilldata/web-common/runtime-client/manual-clients";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  export let data: { fileArtifact?: FileArtifact } = {};

  const DEFAULT_EDITOR_WIDTH = 400;
  let containerWidth: number;
  let editorWidth = DEFAULT_EDITOR_WIDTH;

  let filePath: string;
  let chartName: string;

  if (data.fileArtifact) {
    filePath = data.fileArtifact.path;
    chartName = data.fileArtifact.getEntityName();
  } else {
    // needed for backwards compatibility for now
    chartName = $page.params.name;
    filePath = getFileAPIPathFromNameAndType(chartName, EntityType.Chart);
  }

  $: fileQuery = createRuntimeServiceGetFile($runtime.instanceId, filePath, {
    query: {
      onError: (err) => {
        if (err.response?.data?.message.includes(CATALOG_ENTRY_NOT_FOUND)) {
          throw error(404, "Dashboard not found");
        }

        throw error(err.response?.status || 500, err.message);
      },
      refetchOnWindowFocus: false,
    },
  });

  $: yaml = $fileQuery.data?.blob || "";

  $: resourceQuery = useResource("default", chartName, ResourceKind.Component);

  $: ({ data: componentResource } = $resourceQuery);

  $: ({ resolverProperties } = componentResource?.component?.spec ?? {});

  $: chartDataQuery = createRuntimeServiceGetChartData(
    queryClient,
    $runtime.instanceId,
    chartName,
    resolverProperties,
  );

  $: chartData = $chartDataQuery?.data as Record<string, unknown>[];
</script>

<svelte:head>
  <title>Rill Developer | {chartName}</title>
</svelte:head>

{#if $fileQuery.data && yaml !== undefined}
  <WorkspaceContainer inspector={false} bind:width={containerWidth}>
    <ChartsHeader slot="header" {filePath} />
    <div slot="body" class="flex size-full">
      <div style:width="{editorWidth}px" class="relative flex-none">
        <Resizer
          direction="EW"
          side="right"
          bind:dimension={editorWidth}
          min={300}
          max={0.6 * containerWidth}
        />
        <ChartsEditor {filePath} />
      </div>
      <div class="size-full flex-col flex">
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

        <div class="w-full h-48 bg-white border-t">
          {#if chartData}
            <PreviewTable
              rows={chartData}
              name={chartName}
              columnNames={Object.keys(chartData[0]).map((key) => ({
                type: "VARCHAR",
                name: key,
              }))}
            />
          {/if}
        </div>
      </div>
    </div>
  </WorkspaceContainer>
{/if}
