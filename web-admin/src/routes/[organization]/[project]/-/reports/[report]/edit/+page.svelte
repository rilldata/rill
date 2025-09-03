<script lang="ts">
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";
  import DashboardStateManager from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import PivotOnlyDashboard from "@rilldata/web-common/features/dashboards/workspace/PivotOnlyDashboard.svelte";
  import { mapQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-to-explore.ts";
  import ScheduledReportSidebar from "@rilldata/web-common/features/scheduled-reports/ScheduledReportSidebar.svelte";
  import { derived } from "svelte/store";
  import type { PageData } from "./$types";

  export let data: PageData;
  const {
    report: reportResource,
    organization,
    project,
    reportName,
    exploreName,
  } = data;

  const dashboardStateForReport = mapQueryToDashboard({
    exploreName,
    queryName: reportResource?.report?.spec?.queryName,
    queryArgsJson: reportResource?.report?.spec?.queryArgsJson,
    annotations: reportResource?.report?.spec?.annotations ?? {},
    forceOpenPivot: true,
  });

  $: metricsViewName = $dashboardStateForReport?.data?.metricsViewName ?? "";
  $: stateStore = derived(dashboardStateForReport, (stateForReport) => ({
    data: stateForReport?.data?.exploreState,
    error: stateForReport?.error,
    isLoading: stateForReport?.isLoading ?? true,
    isFetching: stateForReport?.isFetching ?? true,
  }));
</script>

{#if $dashboardStateForReport?.isLoading}
  Loading...
{:else if $dashboardStateForReport?.error}
  Error: {$dashboardStateForReport?.error}
{:else}
  <StateManagersProvider {metricsViewName} {exploreName}>
    <DashboardStateManager
      {exploreName}
      bookmarkOrTokenExploreState={stateStore}
      disableSessionStorage
      disableMostRecentDashboardState
    >
      <DashboardThemeProvider>
        <div class="flex flex-row h-full">
          <PivotOnlyDashboard {metricsViewName} {exploreName} />
          <ScheduledReportSidebar
            {organization}
            {project}
            props={{
              mode: "edit",
              reportName,
              reportSpec: reportResource?.report?.spec ?? {},
            }}
          />
        </div>
      </DashboardThemeProvider>
    </DashboardStateManager>
  </StateManagersProvider>
{/if}
