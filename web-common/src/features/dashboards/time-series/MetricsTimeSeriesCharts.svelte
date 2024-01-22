<script lang="ts">
  import { runtime } from "../../../runtime-client/runtime-store";
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import { bisectData } from "@rilldata/web-common/components/data-graphic/utils";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import SeachableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SeachableFilterButton.svelte";
  import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors";
  import { createShowHideMeasuresStore } from "@rilldata/web-common/features/dashboards/show-hide-selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { chartInteractionColumn } from "@rilldata/web-common/features/dashboards/time-dimension-details/time-dimension-data-store";
  import BackToOverview from "@rilldata/web-common/features/dashboards/time-series/BackToOverview.svelte";
  import { useTimeSeriesDataStore } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
  import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { getAdjustedChartTime } from "@rilldata/web-common/lib/time/ranges";
  import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
  import Spinner from "../../entity-management/Spinner.svelte";
  import MeasureBigNumber from "../big-number/MeasureBigNumber.svelte";
  import MeasureChart from "./MeasureChart.svelte";
  import MeasureZoom from "./MeasureZoom.svelte";
  import type { DimensionDataItem } from "./multiple-dimension-queries";
  import TimeSeriesChartContainer from "./TimeSeriesChartContainer.svelte";

  export let metricViewName;
  export let workspaceWidth: number;

  $: dashboardStore = useDashboardStore(metricViewName);
  $: instanceId = $runtime.instanceId;

  // query the `/meta` endpoint to get the measures and the default time grain
  $: metaQuery = useMetaQuery(instanceId, metricViewName);

  const {
    selectors: {
      measures: { isMeasureValidPercentOfTotal },
      dimensionFilters: { includedDimensionValues },
    },
  } = getStateManagers();

  $: showHideMeasures = createShowHideMeasuresStore(metricViewName, metaQuery);

  const timeControlsStore = useTimeControlStore(getStateManagers());
  const timeSeriesDataStore = useTimeSeriesDataStore(getStateManagers());

  $: expandedMeasureName = $dashboardStore?.expandedMeasureName;
  $: comparisonDimension = $dashboardStore?.selectedComparisonDimension;
  $: showComparison = !comparisonDimension && $timeControlsStore.showComparison;
  $: interval =
    $timeControlsStore.selectedTimeRange?.interval ??
    $timeControlsStore.minTimeGrain;
  $: isScrubbing = $dashboardStore?.selectedScrubRange?.isScrubbing;

  $: isPercOfTotalAsContextColumn =
    $dashboardStore?.leaderboardContextColumn ===
    LeaderboardContextColumn.PERCENT;
  $: includedValuesForDimension = $includedDimensionValues(
    comparisonDimension as string,
  );

  // List of measures which will be shown on the dashboard
  let renderedMeasures = [];
  $: {
    if (expandedMeasureName) {
      renderedMeasures = $metaQuery.data?.measures.filter(
        (measure) => measure.name === expandedMeasureName,
      );
    } else {
      renderedMeasures = $metaQuery.data?.measures.filter(
        (_, i) => $showHideMeasures.selectedItems[i],
      );
    }
  }

  $: totals = $timeSeriesDataStore.total;
  $: totalsComparisons = $timeSeriesDataStore.comparisonTotal;

  let scrubStart;
  let scrubEnd;

  let mouseoverValue = undefined;
  let startValue: Date;
  let endValue: Date;

  // When changing the timeseries query and the cache is empty, $timeSeriesQuery.data?.data is
  // temporarily undefined as results are fetched.
  // To avoid unmounting TimeSeriesBody, which would cause us to lose our tween animations,
  // we make a copy of the data that avoids `undefined` transition states.
  // TODO: instead, try using svelte-query's `keepPreviousData = True` option.

  let dataCopy;
  let dimensionDataCopy: DimensionDataItem[] = [];
  $: if ($timeSeriesDataStore?.timeSeriesData) {
    dataCopy = $timeSeriesDataStore.timeSeriesData;
  }
  $: formattedData = dataCopy;

  $: if (
    $timeSeriesDataStore?.dimensionChartData?.length ||
    !comparisonDimension ||
    includedValuesForDimension.length === 0
  ) {
    dimensionDataCopy = $timeSeriesDataStore.dimensionChartData || [];
  }
  $: dimensionData = dimensionDataCopy;

  // FIXME: move this logic to a function + write tests.
  $: if ($timeControlsStore.ready) {
    // adjust scrub values for Javascript's timezone changes
    scrubStart = adjustOffsetForZone(
      $dashboardStore?.selectedScrubRange?.start,
      $dashboardStore?.selectedTimezone,
    );
    scrubEnd = adjustOffsetForZone(
      $dashboardStore?.selectedScrubRange?.end,
      $dashboardStore?.selectedTimezone,
    );

    const slicedData =
      $timeControlsStore.selectedTimeRange?.name === TimeRangePreset.ALL_TIME
        ? formattedData?.slice(1)
        : formattedData?.slice(1, -1);
    chartInteractionColumn.update((state) => {
      const { start, end } = getOrderedStartEnd(scrubStart, scrubEnd);

      const startPos = bisectData(
        start,
        "center",
        "ts_position",
        slicedData,
        true,
      );
      const endPos = bisectData(end, "center", "ts_position", slicedData, true);

      return {
        hover: isScrubbing ? undefined : state.hover,
        scrubStart: startPos,
        scrubEnd: endPos,
      };
    });

    const adjustedChartValue = getAdjustedChartTime(
      $timeControlsStore.selectedTimeRange?.start,
      $timeControlsStore.selectedTimeRange?.end,
      $dashboardStore?.selectedTimezone,
      interval,
      $timeControlsStore.selectedTimeRange?.name,
      $metaQuery.data.defaultTimeRange,
    );

    startValue = adjustedChartValue?.start;
    endValue = adjustedChartValue?.end;
  }

  $: if (
    expandedMeasureName &&
    formattedData &&
    $timeControlsStore.selectedTimeRange &&
    !isScrubbing
  ) {
    if (!mouseoverValue?.x) {
      chartInteractionColumn.update((state) => ({
        ...state,
        hover: undefined,
      }));
    } else {
      const columnNum = bisectData(
        mouseoverValue.x,
        "center",
        "ts_position",
        $timeControlsStore.selectedTimeRange?.name === TimeRangePreset.ALL_TIME
          ? formattedData?.slice(1)
          : formattedData?.slice(1, -1),
        true,
      );

      if ($chartInteractionColumn?.hover !== columnNum)
        chartInteractionColumn.update((state) => ({
          ...state,
          hover: columnNum,
        }));
    }
  }

  const toggleMeasureVisibility = (e) => {
    showHideMeasures.toggleVisibility(e.detail.name);
  };
  const setAllMeasuresNotVisible = () => {
    showHideMeasures.setAllToNotVisible();
  };
  const setAllMeasuresVisible = () => {
    showHideMeasures.setAllToVisible();
  };
</script>

<TimeSeriesChartContainer
  enableFullWidth={Boolean(expandedMeasureName)}
  end={endValue}
  start={startValue}
  {workspaceWidth}
>
  <div class="bg-white sticky top-0 flex pl-1 z-10">
    {#if expandedMeasureName}
      <BackToOverview {metricViewName} />
    {:else}
      <SeachableFilterButton
        label="Measures"
        on:deselect-all={setAllMeasuresNotVisible}
        on:item-clicked={toggleMeasureVisibility}
        on:select-all={setAllMeasuresVisible}
        selectableItems={$showHideMeasures.selectableItems}
        selectedItems={$showHideMeasures.selectedItems}
        tooltipText="Choose measures to display"
      />
    {/if}
  </div>
  <div class="bg-white sticky left-0 top-0 overflow-visible z-10">
    <!-- top axis element -->
    <div />
    <MeasureZoom {metricViewName} />
    {#if $dashboardStore?.selectedTimeRange && startValue && endValue}
      <SimpleDataGraphic
        height={26}
        overflowHidden={false}
        top={29}
        bottom={0}
        xMin={startValue}
        xMax={endValue}
      >
        <Axis superlabel side="top" placement="start" />
      </SimpleDataGraphic>
    {/if}
  </div>
  <!-- bignumbers and line charts -->
  {#if renderedMeasures.length}
    <!-- FIXME: this is pending the remaining state work for show/hide measures and dimensions -->
    {#each renderedMeasures as measure (measure.name)}
      <!-- FIXME: I can't select the big number by the measure id. -->
      <!-- for bigNum, catch nulls and convert to undefined.  -->
      {@const bigNum = totals?.[measure.name] ?? undefined}
      {@const comparisonValue = totalsComparisons?.[measure.name]}
      {@const isValidPercTotal = $isMeasureValidPercentOfTotal(measure.name)}
      {@const comparisonPercChange =
        comparisonValue && bigNum !== undefined && bigNum !== null
          ? (bigNum - comparisonValue) / comparisonValue
          : undefined}
      <MeasureBigNumber
        {measure}
        value={bigNum}
        isMeasureExpanded={!!expandedMeasureName}
        {showComparison}
        comparisonOption={$timeControlsStore?.selectedComparisonTimeRange?.name}
        {comparisonValue}
        {comparisonPercChange}
        status={$timeSeriesDataStore?.isFetching
          ? EntityStatus.Running
          : EntityStatus.Idle}
        on:expand-measure={() => {
          metricsExplorerStore.setExpandedMeasureName(
            metricViewName,
            measure.name,
          );
        }}
      />

      <div style:height="125px">
        {#if $timeSeriesDataStore?.isError}
          <div class="p-5"><CrossIcon /></div>
        {:else if formattedData}
          <MeasureChart
            bind:mouseoverValue
            {measure}
            {isScrubbing}
            {scrubStart}
            {scrubEnd}
            {metricViewName}
            data={formattedData}
            {dimensionData}
            zone={$dashboardStore?.selectedTimezone}
            xAccessor="ts_position"
            labelAccessor="ts"
            timeGrain={interval}
            yAccessor={measure.name}
            xMin={startValue}
            xMax={endValue}
            {showComparison}
            validPercTotal={isPercOfTotalAsContextColumn && isValidPercTotal
              ? bigNum
              : null}
            mouseoverTimeFormat={(value) => {
              /** format the date according to the time grain */
              return new Date(value).toLocaleDateString(
                undefined,
                TIME_GRAIN[interval].formatDate,
              );
            }}
          />
        {:else}
          <div class="flex items-center justify-center w-24">
            <Spinner status={EntityStatus.Running} />
          </div>
        {/if}
      </div>
    {/each}
  {/if}
</TimeSeriesChartContainer>
