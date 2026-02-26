<script lang="ts">
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateManager from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import { derived } from "svelte/store";
  import { isHTTPError } from "@rilldata/web-common/lib/errors";
  import { createRuntimeServiceGetExplore } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { errorStore } from "../../components/errors/error-store";
  import { EmbedStorageNamespacePrefix } from "@rilldata/web-admin/features/embeds/constants.ts";
  import {
    getEmbedThemeStoreInstance,
    resolveEmbedTheme,
  } from "@rilldata/web-common/features/embeds/embed-theme";

  export let exploreName: string;

  const runtimeClient = useRuntimeClient();

  $: explore = createRuntimeServiceGetExplore(runtimeClient, {
    name: exploreName,
  });
  $: ({ isSuccess, isError, error, data } = $explore);
  $: isExploreNotFound =
    isError && isHTTPError(error) && error.response.status === 404;

  // We check for explore.state.validSpec instead of meta.reconcileError. validSpec persists
  // from previous valid explores, allowing display even when the current explore spec is invalid
  // and a meta.reconcileError exists.
  $: isExploreErrored = !data?.explore?.explore?.state?.validSpec;

  $: metricsViewName = data?.metricsView?.meta?.name?.name;

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
