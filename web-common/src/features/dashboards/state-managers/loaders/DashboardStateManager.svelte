<script lang="ts">
  import { afterNavigate, onNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    ExploreUrlLimitWarningBannerID,
    ExploreUrlLimitWarningBannerPriority,
  } from "@rilldata/web-common/components/banner/constants";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
  import { DashboardStateDataLoader } from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateDataLoader";
  import { DashboardStateSync } from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateSync";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
  import DashboardLoading from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardLoading.svelte";
  import { isUrlTooLong } from "@rilldata/web-common/features/dashboards/url-state/url-length-limits";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    extractErrorMessage,
    extractErrorStatusCode,
  } from "@rilldata/web-common/lib/errors";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { onDestroy } from "svelte";

  export let exploreName: string;
  export let storageNamespacePrefix: string | undefined = undefined;
  export let bookmarkOrTokenExploreState:
    | CompoundQueryResult<Partial<ExploreState> | null>
    | undefined = undefined;
  export let disableMostRecentDashboardState: boolean = false;

  const client = useRuntimeClient();

  $: exploreSpecQuery = useExploreValidSpec(client, exploreName);
  $: exploreSpec = $exploreSpecQuery.data?.explore ?? {};
  $: metricsViewName = exploreSpec?.metricsView ?? "";
  $: exploreStore = useExploreState(exploreName);

  $: dataLoader = new DashboardStateDataLoader(
    client,
    exploreName,
    storageNamespacePrefix,
    bookmarkOrTokenExploreState,
    disableMostRecentDashboardState,
  );

  let stateSync: DashboardStateSync | undefined;
  $: if (dataLoader) {
    stateSync?.teardown();
    stateSync = new DashboardStateSync(
      client,
      metricsViewName,
      exploreName,
      storageNamespacePrefix,
      dataLoader,
    );
  }

  let initExploreState:
    | CompoundQueryResult<ExploreState | undefined>
    | undefined;
  $: if (dataLoader) ({ initExploreState } = dataLoader);

  let error: Error | null;
  let isLoading: boolean;
  $: if (initExploreState) {
    ({ isLoading, error } = $initExploreState as {
      isLoading: boolean;
      error: Error | null;
    });
  }

  $: showUrlWarning = isUrlTooLong($page.url);
  $: if (showUrlWarning) {
    eventBus.emit("add-banner", {
      id: ExploreUrlLimitWarningBannerID,
      priority: ExploreUrlLimitWarningBannerPriority,
      message: {
        type: "warning",
        message:
          "URL is too long. Some features like export will not work. Please remove some filters.",
        iconType: "alert",
      },
    });
  } else {
    eventBus.emit("remove-banner", ExploreUrlLimitWarningBannerID);
  }

  afterNavigate(({ from, to, type }) => {
    if (!from?.url || !to?.url || !stateSync) return;

    void stateSync.handleURLChange(to.url.searchParams, type);
  });

  onNavigate(({ from, to }) => {
    const changedDashboard =
      !from || !to || from.params?.dashboard !== to.params?.dashboard;
    // Clear out any dashboard banners
    // Note: we still have this on top of the above reactive statement to handle cases where navigation is to a non-dashboard route.
    if (changedDashboard) {
      eventBus.emit("remove-banner", ExploreUrlLimitWarningBannerID);
    }
  });

  onDestroy(() => {
    stateSync?.teardown();
  });
</script>

{#if isLoading}
  <DashboardLoading {isLoading} />
{:else if error}
  <ErrorPage
    statusCode={extractErrorStatusCode(error)}
    header="Failed to load dashboard"
    detail={extractErrorMessage(error)}
  />
{:else if $exploreStore}
  <slot />
{/if}
