<script lang="ts">
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardBuilding from "@rilldata/web-common/features/dashboards/DashboardBuilding.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateManager from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import {
    useExploreWithPolling,
    isExploreReconcilingForFirstTime,
    isExploreErrored,
  } from "@rilldata/web-common/features/explores/selectors";
  import { derived } from "svelte/store";
  import { isNotFoundError } from "@rilldata/web-common/lib/errors";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { errorStore } from "../../components/errors/error-store";
  import { EmbedStorageNamespacePrefix } from "@rilldata/web-admin/features/embeds/constants.ts";
  import {
    getEmbedThemeStoreInstance,
    resolveEmbedTheme,
  } from "@rilldata/web-common/features/embeds/embed-theme";

  export let exploreName: string;

  const runtimeClient = useRuntimeClient();

  $: explore = useExploreWithPolling(runtimeClient, exploreName);
  $: isExploreNotFound =
    !$explore.data && $explore.isError && isNotFoundError($explore.error);

  $: metricsViewName = $explore.data?.metricsView?.meta?.name?.name;

  const embedThemeStore = getEmbedThemeStoreInstance();
  const embedResolvedTheme = derived([embedThemeStore], () =>
    resolveEmbedTheme(),
  );

  // If no dashboard is found, show a 404 page
  $: if (isExploreNotFound) {
    errorStore.set({
      statusCode: 404,
      header: "Explore not found",
      body: `The Explore dashboard you requested could not be found. Please check that you provided the name of a working dashboard.`,
    });
  }
</script>

{#if $explore.isSuccess}
  {#if isExploreReconcilingForFirstTime($explore.data)}
    <DashboardBuilding />
  {:else if isExploreErrored($explore.data)}
    <br /> Explore Error <br />
  {:else if metricsViewName}
    {#key exploreName}
      <StateManagersProvider {exploreName} {metricsViewName}>
        <DashboardStateManager
          {exploreName}
          storageNamespacePrefix={EmbedStorageNamespacePrefix}
          disableMostRecentDashboardState
        >
          <Dashboard
            {exploreName}
            {metricsViewName}
            isEmbedded
            embedThemeName={embedResolvedTheme}
          />
        </DashboardStateManager>
      </StateManagersProvider>
    {/key}
  {/if}
{/if}
