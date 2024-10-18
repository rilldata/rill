<script lang="ts">
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";
  import { resetSelectedMockUserAfterNavigate } from "@rilldata/web-common/features/dashboards/granular-access-policies/resetSelectedMockUserAfterNavigate";
  import { selectedMockUserStore } from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
  import DashboardURLStateSync from "@rilldata/web-common/features/dashboards/url-state/DashboardURLStateSync.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/stores/DashboardStateProvider.svelte";
  import { useProjectParser } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import type { PageData } from "./$types";

  const queryClient = useQueryClient();

  export let data: PageData;
  $: ({ metricsView, explore, partialMetrics } = data);

  resetSelectedMockUserAfterNavigate(queryClient);

  $: exploreName = explore.meta?.name?.name as string;
  $: metricsViewName = metricsView?.meta?.name?.name as string;

  $: ({ instanceId } = $runtime);

  $: filePaths = [
    ...(explore.meta?.filePaths ?? []),
    ...(metricsView.meta?.filePaths ?? []),
  ];
  $: exploreQuery = useExploreValidSpec(instanceId, exploreName);
  $: measures = $exploreQuery.data?.explore?.measures ?? [];
  $: projectParserQuery = useProjectParser(queryClient, instanceId, {
    enabled: $selectedMockUserStore?.admin,
  });

  $: dashboardFileHasParseError =
    $projectParserQuery.data?.projectParser?.state?.parseErrors?.filter(
      (error) => filePaths.includes(error.filePath as string),
    );
  $: mockUserHasNoAccess =
    $selectedMockUserStore && $exploreQuery.error?.response?.status === 404;

  console.log("???");
</script>

<svelte:head>
  <title>Rill Developer | {exploreName}</title>
</svelte:head>

{#if measures.length === 0 && $selectedMockUserStore !== null}
  <ErrorPage
    statusCode={$exploreQuery.error?.response?.status}
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
    statusCode={$exploreQuery.error?.response?.status}
    header="This user can't access this dashboard"
    body="The security policy for this dashboard may make contents invisible to you. If you deploy this dashboard, {$selectedMockUserStore?.email} will see a 404."
  />
{:else}
  {#key exploreName}
    <StateManagersProvider {metricsViewName} {exploreName}>
      <DashboardStateProvider {exploreName}>
        <DashboardURLStateSync {partialMetrics}>
          <DashboardThemeProvider>
            <Dashboard {metricsViewName} {exploreName} />
          </DashboardThemeProvider>
        </DashboardURLStateSync>
      </DashboardStateProvider>
    </StateManagersProvider>
  {/key}
{/if}
