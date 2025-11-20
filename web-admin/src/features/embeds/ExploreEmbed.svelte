<script lang="ts">
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateManager from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import { createRuntimeServiceGetExplore } from "@rilldata/web-common/runtime-client";
  import { errorStore } from "../../components/errors/error-store";
  import { EmbedStorageNamespacePrefix } from "@rilldata/web-admin/features/embeds/constants.ts";

  export let instanceId: string;
  export let exploreName: string;

  $: explore = createRuntimeServiceGetExplore(instanceId, {
    name: exploreName,
  });
  $: ({ isSuccess, isError, error, data } = $explore);
  $: isExploreNotFound = isError && error?.response?.status === 404;

  // We check for explore.state.validSpec instead of meta.reconcileError. validSpec persists
  // from previous valid explores, allowing display even when the current explore spec is invalid
  // and a meta.reconcileError exists.
  $: isExploreErrored = !data?.explore?.explore?.state?.validSpec;

  $: metricsViewName = data?.metricsView?.meta?.name?.name;

  // If no dashboard is found, show a 404 page
  $: if (isExploreNotFound) {
    errorStore.set({
      statusCode: 404,
      header: "Explore not found",
      body: `The Explore dashboard you requested could not be found. Please check that you provided the name of a working dashboard.`,
    });
  }
</script>

{#if isSuccess}
  {#if isExploreErrored}
    <br /> Explore Error <br />
  {:else if data}
    {#key exploreName}
      <StateManagersProvider {exploreName} {metricsViewName}>
        <DashboardStateManager
          {exploreName}
          storageNamespacePrefix={EmbedStorageNamespacePrefix}
          disableMostRecentDashboardState
        >
          <Dashboard {exploreName} {metricsViewName} isEmbedded />
        </DashboardStateManager>
      </StateManagersProvider>
    {/key}
  {/if}
{/if}
