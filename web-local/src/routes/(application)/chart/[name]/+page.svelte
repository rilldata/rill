<script lang="ts">
  import { page } from "$app/stores";
  import Charts from "@rilldata/web-common/features/charts/Charts.svelte";
  import ChartsHeader from "@rilldata/web-common/features/charts/ChartsHeader.svelte";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { WorkspaceContainer } from "@rilldata/web-common/layout/workspace";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { CATALOG_ENTRY_NOT_FOUND } from "@rilldata/web-local/lib/errors/messages";
  import { error } from "@sveltejs/kit";

  export let data: { fileArtifact?: FileArtifact } = {};

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
</script>

<svelte:head>
  <title>Rill Developer | {chartName}</title>
</svelte:head>

{#if $fileQuery.data && yaml !== undefined}
  <WorkspaceContainer inspector={false}>
    <ChartsHeader slot="header" {filePath} />
    <Charts slot="body" {filePath} />
  </WorkspaceContainer>
{/if}
