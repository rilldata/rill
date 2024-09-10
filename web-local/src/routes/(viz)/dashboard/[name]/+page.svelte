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

  const queryClient = useQueryClient();

  export let data: PageData;

  resetSelectedMockUserAfterNavigate(queryClient);

  $: metricsViewName = data.metricsView.meta?.name?.name as string;

  $: ({ instanceId } = $runtime);

  $: filePaths = data.metricsView.meta?.filePaths as string[];
  $: dashboard = useDashboard(instanceId, metricsViewName);
  $: measures = $dashboard.data?.metricsView?.state?.validSpec?.measures ?? [];
  $: projectParserQuery = useProjectParser(queryClient, instanceId, {
    enabled: $selectedMockUserStore?.admin,
  });

  $: dashboardFileHasParseError =
    $projectParserQuery.data?.projectParser?.state?.parseErrors?.filter(
      (error) => filePaths.includes(error.filePath as string),
    );
  $: mockUserHasNoAccess =
    $selectedMockUserStore && $dashboard.error?.response?.status === 404;
</script>

<svelte:head>
  <title>Rill Developer | {metricsViewName}</title>
</svelte:head>

{#if measures.length === 0 && $selectedMockUserStore !== null}
  <ErrorPage
    statusCode={$dashboard.error?.response?.status}
    header="Error fetching dashboard"
    body="No measures available"
  />
  <!-- Handle errors from dashboard YAML edits from an external IDE -->
{:else if dashboardFileHasParseError && dashboardFileHasParseError.length > 0}
  <ErrorPage
    header="Error parsing dashboard"
    body="Please check your dashboard's YAML file for errors."
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
