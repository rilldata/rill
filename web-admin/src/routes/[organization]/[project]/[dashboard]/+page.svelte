<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import DashboardBookmarksStateProvider from "@rilldata/web-admin/features/dashboards/DashboardBookmarksStateProvider.svelte";
  import ProjectErrored from "@rilldata/web-admin/features/projects/ProjectErrored.svelte";
  import { viewAsUserStore } from "@rilldata/web-admin/features/view-as-user/viewAsUserStore";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";
  import DashboardURLStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardURLStateProvider.svelte";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/stores/DashboardStateProvider.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { errorStore } from "../../../../features/errors/error-store";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { asyncWait } from "@rilldata/web-common/lib/waitUtils";
  import { parseFilterString } from "@rilldata/web-common/lib/url/parsing";

  $: ({ instanceId } = $runtime);

  $: ({
    organization: orgName,
    project: projectName,
    dashboard: dashboardName,
  } = $page.params);

  $: ({ searchParams } = $page.url);

  $: filter = searchParams.get("filter") ?? "";

  $: dimensionNames =
    $dashboard?.data?.metricsView?.state?.validSpec?.dimensions.map(
      (d) => d.name,
    ) ?? [];

  $: ({ initDimensions, errorMessage } = parseFilterString(
    filter,
    dimensionNames,
  ));

  $: if (errorMessage) {
    asyncWait(300)
      .then(() => {
        eventBus.emit("notification", {
          message: errorMessage,
          detail: `Filter: ${filter ?? ""}`,
          type: "error",
          options: { persisted: true },
        });
      })
      .catch((e) => {
        console.error(e);
      });
  }

  const user = createAdminServiceGetCurrentUser();

  $: dashboard = useDashboard(instanceId, dashboardName);
  $: isDashboardNotFound =
    !$dashboard.data &&
    $dashboard.isError &&
    $dashboard.error?.response?.status === 404;
  // We check for metricsView.state.validSpec instead of meta.reconcileError. validSpec persists
  // from previous valid dashboards, allowing display even when the current dashboard spec is invalid
  // and a meta.reconcileError exists.
  $: isDashboardErrored = !$dashboard.data?.metricsView?.state?.validSpec;
  $: metricViewName = $dashboard.data?.meta.name.name;
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
  <title>{dashboardName} - Rill</title>
</svelte:head>

<!-- Note: Project and dashboard states might appear to diverge. A project could be errored 
  because dashboard #1 is errored, but dashboard #2 could be OK.  -->

{#if $dashboard.isSuccess}
  {#if isDashboardErrored}
    <ProjectErrored organization={orgName} project={projectName} />
  {:else if metricViewName}
    {#key metricViewName}
      <StateManagersProvider metricsViewName={metricViewName}>
        {#if $user.isSuccess && $user.data.user}
          <DashboardBookmarksStateProvider {metricViewName}>
            <DashboardURLStateProvider {metricViewName}>
              <DashboardThemeProvider>
                <Dashboard {metricViewName} {initDimensions} />
              </DashboardThemeProvider>
            </DashboardURLStateProvider>
          </DashboardBookmarksStateProvider>
        {:else}
          <DashboardStateProvider {metricViewName}>
            <DashboardURLStateProvider {metricViewName}>
              <DashboardThemeProvider>
                <Dashboard {metricViewName} {initDimensions} />
              </DashboardThemeProvider>
            </DashboardURLStateProvider>
          </DashboardStateProvider>
        {/if}
      </StateManagersProvider>
    {/key}
  {/if}
{/if}
