<script lang="ts">
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";
  import DashboardURLStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardURLStateProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/stores/DashboardStateProvider.svelte";
  import { createRuntimeServiceGetExplore } from "@rilldata/web-common/runtime-client";
  import { errorStore } from "../errors/error-store";

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
        <DashboardStateProvider {exploreName}>
          <DashboardURLStateProvider {metricsViewName}>
            <DashboardThemeProvider>
              <Dashboard {exploreName} {metricsViewName} isEmbedded />
            </DashboardThemeProvider>
          </DashboardURLStateProvider>
        </DashboardStateProvider>
      </StateManagersProvider>
    {/key}
  {/if}
{/if}
