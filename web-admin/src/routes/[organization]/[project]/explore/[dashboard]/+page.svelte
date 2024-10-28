<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import DashboardBookmarksStateProvider from "@rilldata/web-admin/features/dashboards/DashboardBookmarksStateProvider.svelte";
  import DashboardBuilding from "@rilldata/web-admin/features/dashboards/DashboardBuilding.svelte";
  import DashboardErrored from "@rilldata/web-admin/features/dashboards/DashboardErrored.svelte";
  import { errorStore } from "@rilldata/web-admin/features/errors/error-store";
  import { viewAsUserStore } from "@rilldata/web-admin/features/view-as-user/viewAsUserStore";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";
  import DashboardURLStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardURLStateProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/stores/DashboardStateProvider.svelte";
  import { useExplore } from "@rilldata/web-common/features/explores/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  const user = createAdminServiceGetCurrentUser();

  const PollIntervalWhenDashboardFirstReconciling = 1000;
  const PollIntervalWhenDashboardErrored = 5000;
  // const PollIntervalWhenDashboardOk = 60000; // This triggers a layout shift, so removing for now

  $: instanceId = $runtime?.instanceId;

  $: ({
    organization: orgName,
    project: projectName,
    dashboard: exploreName,
  } = $page.params);

  $: explore = useExplore(instanceId, exploreName, {
    refetchInterval: () => {
      if (isDashboardReconcilingForFirstTime) {
        return PollIntervalWhenDashboardFirstReconciling;
      } else if (isDashboardErrored) {
        return PollIntervalWhenDashboardErrored;
      } else {
        return false;
      }
    },
  });
  $: exploreTitle =
    $explore.data?.explore?.explore?.state?.validSpec?.displayName;

  $: isDashboardNotFound =
    !$explore.data &&
    $explore.isError &&
    $explore.error?.response?.status === 404;
  // TODO: should these be checking metricsView or explore?
  $: isDashboardReconcilingForFirstTime =
    $explore?.data?.metricsView?.metricsView?.state?.validSpec === null &&
    !$explore?.data?.metricsView?.meta?.reconcileError;
  // We check for metricsView.state.validSpec instead of meta.reconcileError. validSpec persists
  // from previous valid dashboards, allowing display even when the current dashboard spec is invalid
  // and a meta.reconcileError exists.
  $: isDashboardErrored =
    $explore?.data?.metricsView?.metricsView?.state?.validSpec === null &&
    !!$explore?.data?.metricsView?.meta?.reconcileError;
  $: metricsViewName = $explore.data?.metricsView?.meta?.name?.name;

  // If no dashboard is found, show a 404 page
  $: if (isDashboardNotFound) {
    errorStore.set({
      statusCode: 404,
      header: "Dashboard not found",
      body: `The dashboard you requested could not be found. Please check that you provided the name of a working dashboard.`,
    });
  }

  onNavigate(() => {
    // Temporary: clear the mocked user when navigating away.
    // In the future, we should be able to handle the mocked user on all project pages.
    viewAsUserStore.set(null);
    errorStore.reset();
  });
</script>

<svelte:head>
  <title>{exploreTitle || `${exploreName} - Rill`}</title>
</svelte:head>

{#if $explore.isSuccess}
  {#if isDashboardReconcilingForFirstTime}
    <DashboardBuilding />
  {:else if isDashboardErrored}
    <DashboardErrored organization={orgName} project={projectName} />
  {:else if metricsViewName}
    {#key metricsViewName}
      <StateManagersProvider {metricsViewName} {exploreName}>
        {#if $user.isSuccess && $user.data.user}
          <DashboardBookmarksStateProvider {metricsViewName} {exploreName}>
            <DashboardURLStateProvider {metricsViewName}>
              <DashboardThemeProvider>
                <Dashboard {metricsViewName} {exploreName} />
              </DashboardThemeProvider>
            </DashboardURLStateProvider>
          </DashboardBookmarksStateProvider>
        {:else}
          <DashboardStateProvider {exploreName}>
            <DashboardURLStateProvider {metricsViewName}>
              <DashboardThemeProvider>
                <Dashboard {metricsViewName} {exploreName} />
              </DashboardThemeProvider>
            </DashboardURLStateProvider>
          </DashboardStateProvider>
        {/if}
      </StateManagersProvider>
    {/key}
  {/if}
{/if}
