<script lang="ts">
  import { goto } from "$app/navigation";
  import { getEltSize } from "@rilldata/web-common/features/dashboards/get-element-size";
  import {
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import { getContext } from "svelte";
  import type { Tweened } from "svelte/motion";
  import { runtime } from "../../../runtime-client/runtime-store";
  import MeasuresContainer from "../big-number/MeasuresContainer.svelte";
  import { useDashboardStore } from "../dashboard-stores";
  import DimensionDisplay from "../dimension-table/DimensionDisplay.svelte";
  import Filters from "../filters/Filters.svelte";
  import LeaderboardDisplay from "../leaderboard/LeaderboardDisplay.svelte";
  import RowsViewerAccordion from "../rows-viewer/RowsViewerAccordion.svelte";
  import TimeControls from "../time-controls/TimeControls.svelte";
  import MetricsTimeSeriesCharts from "../time-series/MetricsTimeSeriesCharts.svelte";
  import DashboardCTAs from "./DashboardCTAs.svelte";
  import DashboardTitle from "./DashboardTitle.svelte";

  export let metricViewName: string;

  export let leftMargin = undefined;

  $: metricsViewQuery = useMetaQuery($runtime.instanceId, metricViewName);
  $: if ($metricsViewQuery.data) {
    if (!$featureFlags.readOnly && !$metricsViewQuery.data?.measures?.length) {
      goto(`/dashboard/${metricViewName}/edit`);
    }
  }

  $: if (!$featureFlags.readOnly && $metricsViewQuery.isError) {
    goto(`/dashboard/${metricViewName}/edit`);
  }

  let exploreContainerWidth;

  $: metricsExplorer = useDashboardStore(metricViewName);

  $: selectedDimensionName = $metricsExplorer?.selectedDimensionName;
  $: metricTimeSeries = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $metricTimeSeries.data;

  // the navigationVisibilityTween is a tweened value that is used
  // to animate the extra padding that needs to be added to the
  // dashboard container when the navigation pane is collapsed
  const navigationVisibilityTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Tweened<number>;

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();

  $: exploreContainerWidth = getEltSize($observedNode, "x");

  $: leftSide = leftMargin
    ? leftMargin
    : `calc(${$navigationVisibilityTween * 24}px + 1.25rem)`;

  $: isRillDeveloper = $featureFlags.readOnly === false;
</script>

<section
  use:listenToNodeResize
  class="flex flex-col gap-y-1 h-full overflow-x-auto overflow-y-hidden"
>
  <div
    class="border-b mb-3 w-full flex flex-col"
    style:padding-left={leftSide}
    id="header"
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

    <div class="-ml-3 p-1 py-2 space-y-2">
      <TimeControls {metricViewName} />
      {#key metricViewName}
        <Filters />
      {/key}
    </div>
  </div>

  <div
    class="flex gap-x-1 h-full overflow-hidden"
    style:padding-left={leftSide}
  >
    <div class="overflow-y-scroll pb-8 flex-none">
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

    <div class="overflow-y-hidden px-4 grow">
      {#if selectedDimensionName}
        <DimensionDisplay
          {metricViewName}
          dimensionName={selectedDimensionName}
        />
      {:else}
        <LeaderboardDisplay {metricViewName} />
      {/if}
    </div>
  </div>

  {#if isRillDeveloper}
    <RowsViewerAccordion {metricViewName} />
  {/if}
</section>
