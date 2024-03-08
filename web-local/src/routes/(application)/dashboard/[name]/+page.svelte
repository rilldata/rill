<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/stores/DashboardStateProvider.svelte";
  import { resetSelectedMockUserAfterNavigate } from "@rilldata/web-common/features/dashboards/granular-access-policies/resetSelectedMockUserAfterNavigate";
  import { selectedMockUserStore } from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
  import DashboardURLStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardURLStateProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import {
    getResourceStatusStore,
    ResourceStatus,
  } from "@rilldata/web-common/features/entity-management/resource-status-utils";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { WorkspaceContainer } from "@rilldata/web-common/layout/workspace";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { error } from "@sveltejs/kit";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { CATALOG_ENTRY_NOT_FOUND } from "../../../../lib/errors/messages";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";

  const queryClient = useQueryClient();

  const { readOnly } = featureFlags;

  $: metricViewName = $page.params.name;
  $: filePath = getFilePathFromNameAndType(
    metricViewName,
    EntityType.MetricsDefinition,
  );

  $: fileQuery = createRuntimeServiceGetFile($runtime.instanceId, filePath, {
    query: {
      onError: (err) => {
        if (err.response?.data?.message.includes(CATALOG_ENTRY_NOT_FOUND)) {
          throw error(404, "Dashboard not found");
        }

        throw error(err.response?.status || 500, err.message);
      },
    },
  });

  $: resourceStatusStore = getResourceStatusStore(
    queryClient,
    $runtime.instanceId,
    filePath,
    (res) => !!res?.metricsView?.state?.validSpec,
  );
  let showErrorPage = false;
  $: if (metricViewName) {
    showErrorPage = false;
    if ($resourceStatusStore.status === ResourceStatus.Errored) {
      // When the catalog entry doesn't exist, the dashboard config is invalid
      if ($readOnly) {
        throw error(400, "Invalid dashboard");
      }

      // When a mock user doesn't have access to the dashboard, stay on the page to show a message
      if (
        $selectedMockUserStore === null ||
        $resourceStatusStore?.error?.response?.status !== 404
      ) {
        // On all other errors, redirect to the `/edit` page
        goto(`/dashboard/${metricViewName}/edit`);
      } else {
        showErrorPage = true;
      }
    } else if ($resourceStatusStore.status === ResourceStatus.Idle) {
      // Redirect to the `/edit` page if no measures are defined
      if (
        !$readOnly &&
        !$resourceStatusStore.resource?.metricsView?.state?.validSpec?.measures
          ?.length
      ) {
        goto(`/dashboard/${metricViewName}/edit`);
      }
    }
  }

  resetSelectedMockUserAfterNavigate(queryClient);
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

{#if ($fileQuery.data && $resourceStatusStore.status === ResourceStatus.Idle) || showErrorPage}
  <WorkspaceContainer
    assetID={metricViewName}
    bgClass="bg-white"
    inspector={false}
  >
    <svelte:fragment slot="body">
      {#key metricViewName}
        <StateManagersProvider metricsViewName={metricViewName}>
          <DashboardStateProvider {metricViewName}>
            <DashboardURLStateProvider {metricViewName}>
              <DashboardThemeProvider>
                <Dashboard {metricViewName} />
              </DashboardThemeProvider>
            </DashboardURLStateProvider>
          </DashboardStateProvider>
        </StateManagersProvider>
      {/key}
    </svelte:fragment>
  </WorkspaceContainer>
{:else if $resourceStatusStore.status === ResourceStatus.Busy}
  <WorkspaceContainer
    assetID={metricViewName}
    bgClass="bg-white"
    inspector={false}
  >
    <div class="grid h-screen place-content-center" slot="body">
      <ReconcilingSpinner />
    </div>
  </WorkspaceContainer>
{/if}
