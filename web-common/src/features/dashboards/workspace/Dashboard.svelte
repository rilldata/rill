<script lang="ts">
  import { getEltSize } from "@rilldata/web-common/features/dashboards/get-element-size";
  import PivotDisplay from "@rilldata/web-common/features/dashboards/pivot/PivotDisplay.svelte";
  import {
    useDashboard,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import TabBar from "@rilldata/web-common/features/dashboards/tab-bar/TabBar.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import { getContext } from "svelte";
  import type { Tweened } from "svelte/motion";
  import { useDashboardStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { runtime } from "../../../runtime-client/runtime-store";
  import MeasuresContainer from "../big-number/MeasuresContainer.svelte";
  import DimensionDisplay from "../dimension-table/DimensionDisplay.svelte";
  import Filters from "../filters/Filters.svelte";
  import MockUserHasNoAccess from "../granular-access-policies/MockUserHasNoAccess.svelte";
  import { selectedMockUserStore } from "../granular-access-policies/stores";
  import LeaderboardDisplay from "../leaderboard/LeaderboardDisplay.svelte";
  import RowsViewerAccordion from "../rows-viewer/RowsViewerAccordion.svelte";
  import TimeControls from "../time-controls/TimeControls.svelte";
  import TimeDimensionDisplay from "../time-dimension-details/TimeDimensionDisplay.svelte";
  import MetricsTimeSeriesCharts from "../time-series/MetricsTimeSeriesCharts.svelte";
  import DashboardCTAs from "./DashboardCTAs.svelte";
  import DashboardTitle from "./DashboardTitle.svelte";

  export let metricViewName: string;
  export let leftMargin = undefined;

  const { cloudDataViewer } = featureFlags;

  let exploreContainerWidth;

  $: metricsExplorer = useDashboardStore(metricViewName);

  $: selectedDimensionName = $metricsExplorer?.selectedDimensionName;
  $: expandedMeasureName = $metricsExplorer?.expandedMeasureName;
  $: showPivot = $metricsExplorer?.pivot?.active;
  $: metricTimeSeries = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName,
  );
  $: hasTimeSeries = $metricTimeSeries.data;

  // flex-row flex-col
  $: dashboardAlignment = expandedMeasureName ? "col" : "row";

  // the navigationVisibilityTween is a tweened value that is used
  // to animate the extra padding that needs to be added to the
  // dashboard container when the navigation pane is collapsed
  const navigationVisibilityTween = getContext<Tweened<number>>(
    "rill:app:navigation-visibility-tween",
  );

  const { readOnly } = featureFlags;

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();

  $: exploreContainerWidth = getEltSize($observedNode, "x");

  $: leftSide = leftMargin
    ? leftMargin
    : `calc(${$navigationVisibilityTween * 24}px + 1.25rem)`;

  $: isRillDeveloper = $readOnly === false;

  // Check if the mock user (if selected) has access to the dashboard
  $: dashboard = useDashboard($runtime.instanceId, metricViewName);
  $: mockUserHasNoAccess =
    $selectedMockUserStore && $dashboard.error?.response?.status === 404;
</script>

<section
  class="flex flex-col h-screen w-full overflow-hidden dashboard-theme-boundary"
  use:listenToNodeResize
>
  <div
    class="border-b w-full flex flex-col bg-slate-50"
    id="header"
    style:padding-left={leftSide}
  >
    {#if isRillDeveloper}
      <!-- FIXME: adding an -mb-3 fixes the spacing issue incurred by changes to the header 
        to accommodate the cloud dashboard. We should go back and reconcile these headers so we 
        don't need to do this. -->
      <div
        class="flex items-center justify-between -mb-3 w-full pl-1 pr-4"
        style:height="var(--header-height)"
      >
        <DashboardTitle {metricViewName} />
        <DashboardCTAs {metricViewName} />
      </div>
    {/if}

    {#if mockUserHasNoAccess}
      <div class="mb-3" />
    {:else}
      <div class="-ml-3 px-1 pt-2 space-y-2">
        <TimeControls {metricViewName} />

        {#key metricViewName}
          <section class="flex justify-between gap-x-4">
            <Filters />
            <div class="flex flex-col justify-end">
              <TabBar />
            </div>
          </section>
        {/key}
      </div>
    {/if}
  </div>

  {#if mockUserHasNoAccess}
    <MockUserHasNoAccess />
  {:else}
    <div class="h-full overflow-hidden">
      {#if showPivot}
        <div class="overflow-y-hidden flex-1">
          <PivotDisplay />
        </div>
      {:else}
        <div
          style:padding-left={leftSide}
          class="flex gap-x-1 gap-y-4 mt-3 w-full h-full flex-{dashboardAlignment} overflow-hidden"
        >
          <div class="h-full flex-content max-h-fit">
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

          <div class="overflow-y-hidden h-full w-full pb-4">
            {#if expandedMeasureName}
              <TimeDimensionDisplay {metricViewName} />
            {:else if selectedDimensionName}
              <DimensionDisplay />
            {:else}
              <LeaderboardDisplay />
            {/if}
          </div>
        </div>
      {/if}
    </div>

    {#if (isRillDeveloper || $cloudDataViewer) && !expandedMeasureName && !showPivot}
      <RowsViewerAccordion {metricViewName} />
    {/if}
  {/if}
</section>

<style>
  .fixed-metric-height {
    height: 280px;
  }
</style>
