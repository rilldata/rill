<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import DashboardBuilding from "@rilldata/web-admin/features/dashboards/DashboardBuilding.svelte";
  import DashboardErrored from "@rilldata/web-admin/features/dashboards/DashboardErrored.svelte";
  import { viewAsUserStore } from "@rilldata/web-admin/features/view-as-user/viewAsUserStore";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardURLStateSync from "@rilldata/web-common/features/dashboards/url-state/DashboardURLStateSync.svelte";
  import { useExplore } from "@rilldata/web-common/features/explores/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import type { V1GetExploreResponse } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { PageData } from "./$types";

  const PollIntervalWhenDashboardFirstReconciling = 1000;
  const PollIntervalWhenDashboardErrored = 5000;
  // const PollIntervalWhenDashboardOk = 60000; // This triggers a layout shift, so removing for now

  export let data: PageData;
  $: ({
    defaultExplorePreset,
    homeBookmarkExploreState,
    exploreStateFromYAMLConfig,
    partialExploreStateFromUrl,
    exploreStateFromSessionStorage,
    errors,
    exploreName,
  } = data);
  $: if (errors?.length) {
    const _errs = errors;
    setTimeout(() => {
      eventBus.emit("notification", {
        type: "error",
        message: _errs[0].message,
        options: {
          persisted: true,
        },
      });
    }, 100);
  }
  $: ({ instanceId } = $runtime);

  $: ({ organization: orgName, project: projectName } = $page.params);

  $: explore = useExplore(instanceId, exploreName, {
    refetchInterval: (data) => {
      if (!data) {
        return false;
      } else if (isDashboardReconcilingForFirstTime(data)) {
        return PollIntervalWhenDashboardFirstReconciling;
      } else if (isDashboardErrored(data)) {
        return PollIntervalWhenDashboardErrored;
      } else {
        return false;
      }
    },
  });
  $: exploreTitle =
    $explore.data?.explore?.explore?.state?.validSpec?.displayName;

  $: metricsViewName = $explore.data?.metricsView?.meta?.name?.name;

  onNavigate(() => {
    // Temporary: clear the mocked user when navigating away.
    // In the future, we should be able to handle the mocked user on all project pages.
    viewAsUserStore.set(null);
  });

  /**
   * The `isDashboardReconcilingForFirstTime` and `isDashboardErrored` helper functions are intentionally used instead of similarly-named variables.
   * Using variables instead of functions would create a circular dependency that chokes Svelte's reactivity, as the values inside the `useExplore` hook would
   * themselves be derived from the output of the `useExplore` hook.
   */
  function isDashboardReconcilingForFirstTime(
    exploreResponse: V1GetExploreResponse,
  ) {
    if (!exploreResponse) return undefined;
    const isMetricsViewReconcilingForFirstTime =
      !exploreResponse.metricsView?.metricsView?.state?.validSpec &&
      !exploreResponse.metricsView?.meta?.reconcileError;
    const isExploreReconcilingForFirstTime =
      !exploreResponse.explore?.explore?.state?.validSpec &&
      !exploreResponse.explore?.meta?.reconcileError;
    return (
      isMetricsViewReconcilingForFirstTime || isExploreReconcilingForFirstTime
    );
  }

  function isDashboardErrored(exploreResponse: V1GetExploreResponse) {
    if (!exploreResponse) return undefined;
    // We only consider a dashboard errored (from the end-user perspective) when BOTH a reconcile error exists AND a validSpec does not exist.
    // If there's any validSpec (which can persist from a previous, non-current spec), then we serve that version of the dashboard to the user,
    // so the user does not see an error state.
    const isMetricsViewErrored =
      !exploreResponse.metricsView?.metricsView?.state?.validSpec &&
      !!exploreResponse.metricsView?.meta?.reconcileError;
    const isExploreErrored =
      !exploreResponse.explore?.explore?.state?.validSpec &&
      !!exploreResponse.explore?.meta?.reconcileError;
    return isMetricsViewErrored || isExploreErrored;
  }
</script>

<svelte:head>
  <title>{exploreTitle || `${exploreName} - Rill`}</title>
</svelte:head>

{#if $explore.isSuccess}
  {#if isDashboardReconcilingForFirstTime($explore.data)}
    <DashboardBuilding />
  {:else if isDashboardErrored($explore.data)}
    <DashboardErrored organization={orgName} project={projectName} />
  {:else if metricsViewName}
    {#key exploreName}
      <StateManagersProvider {metricsViewName} {exploreName}>
        <DashboardURLStateSync
          {metricsViewName}
          {exploreName}
          extraKeyPrefix={`${orgName}__${projectName}__`}
          {defaultExplorePreset}
          initExploreState={homeBookmarkExploreState}
          {exploreStateFromYAMLConfig}
          {partialExploreStateFromUrl}
          {exploreStateFromSessionStorage}
        >
          <DashboardThemeProvider>
            <Dashboard {metricsViewName} {exploreName} />
          </DashboardThemeProvider>
        </DashboardURLStateSync>
      </StateManagersProvider>
    {/key}
  {/if}
{/if}
