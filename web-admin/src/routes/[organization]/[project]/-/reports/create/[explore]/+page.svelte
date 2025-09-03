<script lang="ts">
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";
  import DashboardStateManager from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import PivotOnlyDashboard from "@rilldata/web-common/features/dashboards/workspace/PivotOnlyDashboard.svelte";
  import { mapQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-to-explore.ts";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import ScheduledReportSidebar from "@rilldata/web-common/features/scheduled-reports/ScheduledReportSidebar.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import { derived } from "svelte/store";
  import type { PageData } from "./$types";

  export let data: PageData;
  const { organization, project, exploreName } = data;

  $: ({ instanceId } = $runtime);

  $: exploreSpecQuery = useExploreValidSpec(instanceId, exploreName);
  $: metricsViewName = $exploreSpecQuery.data?.explore?.metricsView ?? "";
</script>

{#if $exploreSpecQuery?.isLoading}
  Loading...
{:else if $exploreSpecQuery?.error}
  Error: {$exploreSpecQuery?.error}
{:else if metricsViewName && exploreName}
  <StateManagersProvider {metricsViewName} {exploreName}>
    <DashboardStateManager
      {exploreName}
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
              mode: "create",
              exploreName,
              query: { metricsViewAggregationRequest: {} },
            }}
          />
        </div>
      </DashboardThemeProvider>
    </DashboardStateManager>
  </StateManagersProvider>
{/if}
