<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { WorkspaceContainer } from "@rilldata/web-common/layout/workspace";
  import {
    createRuntimeServiceGetCatalogEntry,
    createRuntimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { error } from "@sveltejs/kit";
  import { CATALOG_ENTRY_NOT_FOUND } from "../../../../lib/errors/messages";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardStateProvider.svelte";

  $: metricViewName = $page.params.name;
  $: metricsExplorer = useDashboardStore(metricViewName);

  $: fileQuery = createRuntimeServiceGetFile(
    $runtime.instanceId,
    getFilePathFromNameAndType(metricViewName, EntityType.MetricsDefinition),
    {
      query: {
        onError: (err) => {
          if (err.response?.data?.message.includes(CATALOG_ENTRY_NOT_FOUND)) {
            throw error(404, "Dashboard not found");
          }

          throw error(err.response?.status || 500, err.message);
        },
      },
    }
  );

  $: catalogQuery = createRuntimeServiceGetCatalogEntry(
    $runtime.instanceId,
    metricViewName,
    {
      query: {
        onError: () => {
          if (!metricViewName) return;

          // When the catalog entry doesn't exist, the dashboard config is invalid
          if ($featureFlags.readOnly) {
            throw error(400, "Invalid dashboard");
          }

          goto(`/dashboard/${metricViewName}/edit`);
        },
      },
    }
  );
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

{#if $fileQuery.data && $catalogQuery.data}
  <WorkspaceContainer
    top="0px"
    assetID={metricViewName}
    bgClass="bg-white"
    inspector={false}
  >
    <DashboardStateProvider {metricViewName} slot="body">
      <Dashboard {metricViewName} hasTitle />
    </DashboardStateProvider>
  </WorkspaceContainer>
{/if}
