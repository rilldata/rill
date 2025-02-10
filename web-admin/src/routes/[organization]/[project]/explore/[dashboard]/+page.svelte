<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import DashboardBuilding from "@rilldata/web-admin/features/dashboards/DashboardBuilding.svelte";
  import DashboardErrored from "@rilldata/web-admin/features/dashboards/DashboardErrored.svelte";
  import { errorStore } from "@rilldata/web-admin/features/errors/error-store";
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
        options: { persisted: true },
      });
    }, 100);
  }

  $: ({ instanceId } = $runtime);
  $: ({ organization: orgName, project: projectName } = $page.params);

  $: explore = useExplore(instanceId, exploreName, {
    refetchInterval: (data) => {
      if (!data) return false;
      if (isDashboardReconcilingForFirstTime(data))
        return PollIntervalWhenDashboardFirstReconciling;
      if (isDashboardErrored(data)) return PollIntervalWhenDashboardErrored;
      return false;
    },
  });

  $: exploreTitle =
    $explore.data?.explore?.explore?.state?.validSpec?.displayName;
  $: isDashboardNotFound =
    !$explore.data &&
    $explore.isError &&
    $explore.error?.response?.status === 404;
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

  onNavigate(() => {
    viewAsUserStore.set(null);
    errorStore.reset();

    // Clear out any dashboard banners
    if (hasBanner) {
      eventBus.emit("banner", null);
    }
  });

  function isDashboardReconcilingForFirstTime(
    exploreResponse: V1GetExploreResponse,
  ) {
    if (!exploreResponse) return undefined;
    return (
      !exploreResponse.metricsView?.metricsView?.state?.validSpec &&
      !exploreResponse.metricsView?.meta?.reconcileError
    );
  }

  function isDashboardErrored(exploreResponse: V1GetExploreResponse) {
    if (!exploreResponse) return undefined;
    return (
      !exploreResponse.metricsView?.metricsView?.state?.validSpec &&
      !!exploreResponse.metricsView?.meta?.reconcileError
    );
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
