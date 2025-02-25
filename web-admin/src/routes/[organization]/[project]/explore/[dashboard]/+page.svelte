<script lang="ts">
  import { invalidate, onNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import { errorStore } from "@rilldata/web-admin/components/errors/error-store";
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

  export let data: PageData;
  $: ({
    exploreSpecPromise,
    exploreStatePromise,
    homeBookmarkExploreStatePromise,
    exploreName,
  } = data);

  // TODO
  // $: if (errors?.length) {
  //   const _errs = errors;
  //   setTimeout(() => {
  //     eventBus.emit("notification", {
  //       type: "error",
  //       message: _errs[0].message,
  //       options: { persisted: true },
  //     });
  //   }, 100);
  // }

  $: allDataPromise = Promise.all([
    exploreSpecPromise,
    exploreStatePromise,
    homeBookmarkExploreStatePromise,
  ]);
  $: ({ instanceId } = $runtime);
  $: ({ organization: orgName, project: projectName } = $page.params);

  $: explore = useExplore(instanceId, exploreName, {
    refetchInterval: (data) => {
      if (!data) return false;
      if (isExploreReconcilingForFirstTime(data))
        return PollIntervalWhenDashboardFirstReconciling;
      if (isExploreErrored(data)) return PollIntervalWhenDashboardErrored;
      return false;
    },
  });

  $: isDashboardNotFound =
    !$explore.data &&
    $explore.isError &&
    $explore.error?.response?.status === 404;
  $: exploreTitle =
    $explore.data?.explore?.explore?.state?.validSpec?.displayName;
  $: metricsViewName = $explore.data?.metricsView?.meta?.name?.name;
  $: hasBanner = !!$explore.data?.explore?.explore?.state?.validSpec?.banner;

  // If no dashboard is found, show a 404 page
  $: if (isDashboardNotFound) {
    errorStore.set({
      statusCode: 404,
      header: "Dashboard not found",
      body: `The dashboard you requested could not be found. Please check that you provided the name of a working dashboard.`,
    });
  }

  // Display a dashboard banner
  $: if (hasBanner) {
    eventBus.emit("banner", {
      type: "default",
      message: $explore.data.explore.explore.state.validSpec.banner,
      iconType: "alert",
    });
  }

  onNavigate(({ from, to }) => {
    viewAsUserStore.set(null);
    errorStore.reset();

    const changedDashboard =
      !from || !to || from.params.dashboard !== to.params.dashboard;
    // Clear out any dashboard banners
    if (hasBanner && changedDashboard) {
      eventBus.emit("banner", null);
    }
  });

  function isExploreReconcilingForFirstTime(
    exploreResponse: V1GetExploreResponse,
  ) {
    if (!exploreResponse) return undefined;
    const isExploreReconcilingForFirstTime =
      !exploreResponse.explore?.explore?.state?.validSpec &&
      !exploreResponse.explore?.meta?.reconcileError;
    return isExploreReconcilingForFirstTime;
  }

  function isExploreErrored(exploreResponse: V1GetExploreResponse) {
    if (!exploreResponse) return undefined;
    // We only consider a dashboard errored (from the end-user perspective) when BOTH a reconcile error exists AND a validSpec does not exist.
    // If there's any validSpec (which can persist from a previous, non-current spec), then we serve that version of the dashboard to the user,
    // so the user does not see an error state.
    const isExploreErrored =
      !exploreResponse.explore?.explore?.state?.validSpec &&
      !!exploreResponse.explore?.meta?.reconcileError;
    return isExploreErrored;
  }

  let reconcilingForFirstTime: boolean | undefined;
  $: if ($explore.isSuccess) {
    const newReconcilingForFirstTime = isExploreReconcilingForFirstTime(
      $explore.data,
    );
    // reconcilingForFirstTime means the dashboard is reconciling for the 1st time in the current deployment.
    // a new deployment could change this from false to true
    const reconcilingForFirstTimeChanged =
      reconcilingForFirstTime !== undefined &&
      newReconcilingForFirstTime !== undefined &&
      reconcilingForFirstTime !== newReconcilingForFirstTime;
    if (reconcilingForFirstTimeChanged) {
      void invalidate(`explore:${exploreName}`);
    }
    reconcilingForFirstTime = newReconcilingForFirstTime;
  }
</script>

<svelte:head>
  <title>{exploreTitle || `${exploreName} - Rill`}</title>
</svelte:head>

{#if $explore.isSuccess}
  {#if isExploreReconcilingForFirstTime($explore.data)}
    <DashboardBuilding />
  {:else if isExploreErrored($explore.data)}
    <DashboardErrored organization={orgName} project={projectName} />
  {:else if metricsViewName}
    {#key exploreName}
      {#await allDataPromise}
        Loading...
      {:then values}
        <StateManagersProvider {metricsViewName} {exploreName}>
          <DashboardURLStateSync
            {metricsViewName}
            {exploreName}
            extraKeyPrefix={`${orgName}__${projectName}__`}
            defaultExplorePreset={values[0].defaultExplorePreset}
            initExploreState={values[2]}
            exploreStateFromYAMLConfig={values[0].exploreStateFromYAMLConfig}
            partialExploreStateFromUrl={values[1].partialExploreStateFromUrl}
            exploreStateFromSessionStorage={values[1]
              .exploreStateFromSessionStorage}
          >
            <DashboardThemeProvider>
              <Dashboard {metricsViewName} {exploreName} />
            </DashboardThemeProvider>
          </DashboardURLStateSync>
        </StateManagersProvider>
      {/await}
    {/key}
  {/if}
{/if}
