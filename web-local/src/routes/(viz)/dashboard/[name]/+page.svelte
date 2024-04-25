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

  const queryClient = useQueryClient();

  export let data;

  resetSelectedMockUserAfterNavigate(queryClient);

  $: metricsViewName = data.metricsView.meta?.name?.name as string;

  $: filePaths = data.metricsView.meta?.filePaths as string[];
  $: projectParserQuery = useProjectParser(queryClient, data.instanceId);
  $: dashboardFileHasParseError =
    $projectParserQuery.data?.projectParser?.state?.parseErrors?.filter(
      (error) => filePaths.includes(error.filePath as string),
    );
</script>

<svelte:head>
  <title>Rill Developer | {metricsViewName}</title>
</svelte:head>

<!-- Handle errors from dashboard YAML edits from an external IDE   -->
{#if dashboardFileHasParseError && dashboardFileHasParseError.length > 0}
  <ErrorPage
    header="Error parsing dashboard"
    body="Please check your dashboard's YAML file for errors."
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
