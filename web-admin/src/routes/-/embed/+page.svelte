<script lang="ts">
  import { page } from "$app/stores";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardURLStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardURLStateProvider.svelte";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/stores/DashboardStateProvider.svelte";
  import { errorStore } from "../../../features/errors/error-store";

  const instanceId = $page.url.searchParams.get("instance_id");
  const dashboardName = $page.url.searchParams.get("resource");
  // ignoring state and theme params for now

  $: dashboard = useDashboard(instanceId, dashboardName);
  $: isDashboardNotFound =
    $dashboard.isError && $dashboard.error?.response?.status === 404;
  // We check for metricsView.state.validSpec instead of meta.reconcileError. validSpec persists
  // from previous valid dashboards, allowing display even when the current dashboard spec is invalid
  // and a meta.reconcileError exists.
  $: isDashboardErrored = !$dashboard.data?.metricsView?.state?.validSpec;

  // If no dashboard is found, show a 404 page
  $: if (isDashboardNotFound) {
    errorStore.set({
      statusCode: 404,
      header: "Dashboard not found",
      body: `The dashboard you requested could not be found. Please check that you provided the name of a working dashboard.`,
    });
  }
</script>

<svelte:head>
  <title>{dashboardName} - Rill</title>
</svelte:head>

{#if $dashboard.isSuccess}
  {#if isDashboardErrored}
    <br /> Dashboard Error <br />
  {:else}
    {#key dashboardName}
      <StateManagersProvider metricsViewName={dashboardName}>
        <DashboardStateProvider metricViewName={dashboardName}>
          <DashboardURLStateProvider metricViewName={dashboardName}>
            <Dashboard metricViewName={dashboardName} leftMargin={"48px"} />
          </DashboardURLStateProvider>
        </DashboardStateProvider>
      </StateManagersProvider>
    {/key}
  {/if}
{/if}
