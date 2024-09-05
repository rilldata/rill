<script lang="ts">
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";
  import { resetSelectedMockUserAfterNavigate } from "@rilldata/web-common/features/dashboards/granular-access-policies/resetSelectedMockUserAfterNavigate";
  import DashboardURLStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardURLStateProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/stores/DashboardStateProvider.svelte";
  import { useProjectParser } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { useQueryClient } from "@tanstack/svelte-query";
  import type { PageData } from "./$types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import { selectedMockUserStore } from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  const queryClient = useQueryClient();

  export let data: PageData;

  resetSelectedMockUserAfterNavigate(queryClient);

  $: metricsViewName = data.metricsView.meta?.name?.name as string;

  $: ({ instanceId } = $runtime);

  $: filePaths = data.metricsView.meta?.filePaths as string[];
  $: projectParserQuery = useProjectParser(queryClient, instanceId);
  $: dashboardFileHasParseError =
    $projectParserQuery.data?.projectParser?.state?.parseErrors?.filter(
      (error) => filePaths.includes(error.filePath as string),
    );

  $: dashboard = useDashboard(instanceId, metricsViewName);
  $: mockUserHasNoAccess =
    $selectedMockUserStore && $dashboard.error?.response?.status === 404;

  // Handle errors from dashboard YAML edits from an external IDE
  $: if (dashboardFileHasParseError && dashboardFileHasParseError.length > 0) {
    eventBus.emit("banner", {
      type: "error",
      message:
        "Error parsing dashboard – you are viewing your last valid dashboard specification",
      iconType: "alert",
    });
  }

  // TODO: where to get measures from?
  $: hasMeasures = false;

  $: if (mockUserHasNoAccess || !hasMeasures) {
    eventBus.emit("banner", null);
  }
</script>

<svelte:head>
  <title>Rill Developer | {metricsViewName}</title>
</svelte:head>

{#if !hasMeasures && $selectedMockUserStore !== null}
  <ErrorPage
    statusCode={$dashboard.error?.response?.status}
    header="Error fetching dashboard"
    body="No measures available"
  />
{:else if mockUserHasNoAccess}
  <ErrorPage
    statusCode={$dashboard.error?.response?.status}
    header="This user can't access this dashboard"
    body="The security policy for this dashboard may make contents invisible to you. If you deploy this dashboard, {$selectedMockUserStore?.email} will see a 404."
  />
{:else}
  {#key metricsViewName}
    <StateManagersProvider {metricsViewName}>
      <DashboardStateProvider metricViewName={metricsViewName}>
        <DashboardURLStateProvider metricViewName={metricsViewName}>
          <DashboardThemeProvider>
            <Dashboard metricViewName={metricsViewName} />
          </DashboardThemeProvider>
        </DashboardURLStateProvider>
      </DashboardStateProvider>
    </StateManagersProvider>
  {/key}
{/if}
