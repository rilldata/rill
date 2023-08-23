<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/DashboardStateProvider.svelte";
  import { viewAsStore } from "@rilldata/web-common/features/dashboards/granular-access-policies/viewAs";
  import DashboardURLStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardURLStateProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
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

  $: metricViewName = $page.params.name;

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
        onSuccess: (data) => {
          // Redirect to the `/edit` page if no measures are defined
          if (
            !$featureFlags.readOnly &&
            !data.entry.metricsView.measures?.length
          ) {
            goto(`/dashboard/${metricViewName}/edit`);
          }
        },
        onError: (err) => {
          if (!metricViewName) return;

          // When the catalog entry doesn't exist, the dashboard config is invalid
          if ($featureFlags.readOnly) {
            throw error(400, "Invalid dashboard");
          }

          if ($viewAsStore !== null && err.response?.status === 401) {
            // TODO: show a 404 page
            console.log("err", err);
            return;
          }

          // On all other errors, redirect to the `/edit` page
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
    <StateManagersProvider metricsViewName={metricViewName} slot="body">
      {#key metricViewName}
        <DashboardStateProvider {metricViewName}>
          <DashboardURLStateProvider {metricViewName}>
            <Dashboard {metricViewName} hasTitle />
          </DashboardURLStateProvider>
        </DashboardStateProvider>
      {/key}
    </StateManagersProvider>
  </WorkspaceContainer>
{/if}
