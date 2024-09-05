<script lang="ts">
  import { page } from "$app/stores";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import PivotDisplay from "@rilldata/web-common/features/dashboards/pivot/PivotDisplay.svelte";
  import {
    useDashboard,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import TabBar from "@rilldata/web-common/features/dashboards/tab-bar/TabBar.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { navigationOpen } from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import { useDashboardStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { runtime } from "../../../runtime-client/runtime-store";
  import MeasuresContainer from "../big-number/MeasuresContainer.svelte";
  import DimensionDisplay from "../dimension-table/DimensionDisplay.svelte";
  import Filters from "../filters/Filters.svelte";
  import { selectedMockUserStore } from "../granular-access-policies/stores";
  import LeaderboardDisplay from "../leaderboard/LeaderboardDisplay.svelte";
  import RowsViewerAccordion from "../rows-viewer/RowsViewerAccordion.svelte";
  import TimeDimensionDisplay from "../time-dimension-details/TimeDimensionDisplay.svelte";
  import MetricsTimeSeriesCharts from "../time-series/MetricsTimeSeriesCharts.svelte";
  import {
    ResourceKind,
    SingletonProjectParserName,
  } from "../../entity-management/resource-selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { createRuntimeServiceGetResource } from "@rilldata/web-common/runtime-client";

  export let metricViewName: string;

  const { cloudDataViewer, readOnly } = featureFlags;

  let exploreContainerWidth: number;

  $: extraLeftPadding = !$navigationOpen;

  $: metricsExplorer = useDashboardStore(metricViewName);

  $: selectedDimensionName = $metricsExplorer?.selectedDimensionName;
  $: expandedMeasureName = $metricsExplorer?.tdd?.expandedMeasureName;
  $: showPivot = $metricsExplorer?.pivot?.active;
  $: metricTimeSeries = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName,
  );
  $: hasTimeSeries = $metricTimeSeries.data;

  $: isRillDeveloper = $readOnly === false;

  // TODO: or useProjectParser
  $: projectParserQuery = createRuntimeServiceGetResource($runtime.instanceId, {
    "name.kind": ResourceKind.ProjectParser,
    "name.name": SingletonProjectParserName,
  });
  $: ({ data, error } = $projectParserQuery);
  $: parseErrors = data?.resource?.projectParser?.state?.parseErrors;

  // TODO: Only for project admins do we show the error banner (and request the project parser)
  $: console.log("error:", error?.response.data.message);

  // TODO: is it manageProject?
  const canManageProject = false;
  $: if (canManageProject && error?.response.data.message) {
    eventBus.emit("banner", {
      type: "error",
      message: error?.response.data.message ?? parseErrors,
      includesHtml: true,
      iconType: "alert",
    });
  }

  // Check if the mock user (if selected) has access to the dashboard
  $: dashboard = useDashboard($runtime.instanceId, metricViewName);
  $: mockUserHasNoAccess =
    $selectedMockUserStore && $dashboard.error?.response?.status === 404;
</script>

<article
  class="flex flex-col h-screen w-full overflow-y-hidden dashboard-theme-boundary"
  bind:clientWidth={exploreContainerWidth}
>
  <div
    id="header"
    class="border-b w-fit min-w-full flex flex-col bg-slate-50 slide"
    class:left-shift={extraLeftPadding}
  >
    {#if mockUserHasNoAccess}
      <div class="mb-3" />
    {:else}
      {#key metricViewName}
        <section class="flex relative justify-between gap-x-4 py-4 pb-6 px-4">
          <Filters />
          <div class="absolute bottom-0 flex flex-col right-0">
            <TabBar />
          </div>
        </section>
      {/key}
    {/if}
  </div>

  {#if mockUserHasNoAccess}
    <!-- Additional safeguard for mock users without dashboard access. -->
    <ErrorPage
      statusCode={$dashboard.error?.response?.status}
      header="This user can't access this dashboard"
      body="The security policy for this dashboard may make contents invisible to you. If you deploy this dashboard, {$selectedMockUserStore?.email} will see a 404."
    />
  {:else if showPivot}
    <PivotDisplay />
  {:else}
    <div
      class="flex gap-x-1 gap-y-2 size-full overflow-hidden pl-4 slide"
      class:flex-col={expandedMeasureName}
      class:flex-row={!expandedMeasureName}
      class:left-shift={extraLeftPadding}
    >
      <div class="pt-2">
        {#key metricViewName}
          {#if hasTimeSeries}
            <MetricsTimeSeriesCharts
              {metricViewName}
              workspaceWidth={exploreContainerWidth}
            />
          {:else}
            <MeasuresContainer {exploreContainerWidth} {metricViewName} />
          {/if}
        {/key}
      </div>

      {#if expandedMeasureName}
        <hr class="border-t border-gray-200 -ml-4" />
        <TimeDimensionDisplay {metricViewName} />
      {:else if selectedDimensionName}
        <div class="pt-2 pl-1 border-l overflow-auto w-full">
          <DimensionDisplay />
        </div>
      {:else}
        <div class="pt-2 pl-1 border-l overflow-auto w-full">
          <LeaderboardDisplay />
        </div>
      {/if}
    </div>
  {/if}
</article>

{#if (isRillDeveloper || $cloudDataViewer) && !expandedMeasureName && !mockUserHasNoAccess}
  <RowsViewerAccordion {metricViewName} />
{/if}

<style lang="postcss">
  .left-shift {
    @apply pl-8;
  }
</style>
