<script lang="ts">
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import SeachableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SeachableFilterButton.svelte";
  import {
    useDashboardStore,
    metricsExplorerStore,
  } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import {
    humanizeDataType,
    FormatPreset,
    nicelyFormattedTypesToNumberKind,
  } from "@rilldata/web-common/features/dashboards/humanize-numbers";
  import Portal from "@rilldata/web-common/components/Portal.svelte";
  import { FloatingElement } from "@rilldata/web-common/components/floating-element";
  import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors";
  import { createShowHideMeasuresStore } from "@rilldata/web-common/features/dashboards/show-hide-selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { getAdjustedChartTime } from "@rilldata/web-common/lib/time/ranges";
  import {
    createQueryServiceMetricsViewTimeSeries,
    createQueryServiceMetricsViewTotals,
    V1MetricsViewTimeSeriesResponse,
  } from "@rilldata/web-common/runtime-client";
  import type { CreateQueryResult } from "@tanstack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";
  import Spinner from "../../entity-management/Spinner.svelte";
  import MeasureBigNumber from "../big-number/MeasureBigNumber.svelte";
  import MeasureChart from "./MeasureChart.svelte";
  import TimeSeriesChartContainer from "./TimeSeriesChartContainer.svelte";
  import { getOrderedStartEnd, prepareTimeSeries } from "./utils";
  import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
  import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
  import { Button } from "@rilldata/web-common/components/button";
  import Zoom from "@rilldata/web-common/components/icons/Zoom.svelte";

  export let metricViewName;
  export let workspaceWidth: number;

  $: dashboardStore = useDashboardStore(metricViewName);

  $: instanceId = $runtime.instanceId;

  // query the `/meta` endpoint to get the measures and the default time grain
  $: metaQuery = useMetaQuery(instanceId, metricViewName);

  let axisTop;
  const timeControlsStore = useTimeControlStore(getStateManagers());

  $: selectedMeasureNames = $dashboardStore?.selectedMeasureNames;
  $: showComparison = $timeControlsStore.showComparison;
  $: interval =
    $timeControlsStore.selectedTimeRange?.interval ??
    $timeControlsStore.minTimeGrain;

  $: totalsQuery = createQueryServiceMetricsViewTotals(
    $runtime.instanceId,
    metricViewName,
    {
      measureNames: selectedMeasureNames,
      timeStart: $timeControlsStore.timeStart,
      timeEnd: $timeControlsStore.timeEnd,
      filter: $dashboardStore?.filters,
    },
    {
      query: {
        enabled:
          selectedMeasureNames?.length > 0 &&
          $timeControlsStore.ready &&
          !!$dashboardStore?.filters,
      },
    }
  );

  $: totalsComparisonQuery = createQueryServiceMetricsViewTotals(
    instanceId,
    metricViewName,
    {
      measureNames: selectedMeasureNames,
      timeStart: $timeControlsStore.comparisonTimeStart,
      timeEnd: $timeControlsStore.comparisonTimeEnd,
      filter: $dashboardStore?.filters,
    },
    {
      query: {
        enabled: Boolean(showComparison && !!$dashboardStore?.filters),
      },
    }
  );

  // get the totalsComparisons.
  $: totalsComparisons = $totalsComparisonQuery?.data?.data;

  let timeSeriesQuery: CreateQueryResult<
    V1MetricsViewTimeSeriesResponse,
    Error
  >;

  let timeSeriesComparisonQuery: CreateQueryResult<
    V1MetricsViewTimeSeriesResponse,
    Error
  >;

  $: if (
    $dashboardStore &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching &&
    $timeControlsStore.ready
  ) {
    timeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
      instanceId,
      metricViewName,
      {
        measureNames: selectedMeasureNames,
        filter: $dashboardStore?.filters,
        timeStart: $timeControlsStore.adjustedStart,
        timeEnd: $timeControlsStore.adjustedEnd,
        timeGranularity: interval,
        timeZone: $dashboardStore?.selectedTimezone,
      }
    );
    if (showComparison) {
      timeSeriesComparisonQuery = createQueryServiceMetricsViewTimeSeries(
        instanceId,
        metricViewName,
        {
          measureNames: selectedMeasureNames,
          filter: $dashboardStore?.filters,
          timeStart: $timeControlsStore.comparisonAdjustedStart,
          timeEnd: $timeControlsStore.comparisonAdjustedEnd,
          timeGranularity: interval,
          timeZone: $dashboardStore?.selectedTimezone,
        }
      );
    }
  }

  // When changing the timeseries query and the cache is empty, $timeSeriesQuery.data?.data is
  // temporarily undefined as results are fetched.
  // To avoid unmounting TimeSeriesBody, which would cause us to lose our tween animations,
  // we make a copy of the data that avoids `undefined` transition states.
  // TODO: instead, try using svelte-query's `keepPreviousData = True` option.
  let dataCopy;
  let dataComparisonCopy;

  $: if ($timeSeriesQuery?.data?.data) {
    dataCopy = $timeSeriesQuery.data.data;
  }
  $: if ($timeSeriesComparisonQuery?.data?.data)
    dataComparisonCopy = $timeSeriesComparisonQuery.data.data;

  // formattedData adjusts the data to account for Javascript's handling of timezones
  let formattedData;
  let scrubStart;
  let scrubEnd;
  $: if (dataCopy && dataCopy?.length) {
    formattedData = prepareTimeSeries(
      dataCopy,
      dataComparisonCopy,
      TIME_GRAIN[interval].duration,
      $dashboardStore.selectedTimezone
    );

    // adjust scrub values for Javascript's timezone changes
    scrubStart = adjustOffsetForZone(
      $dashboardStore?.selectedScrubRange?.start,
      $dashboardStore?.selectedTimezone
    );
    scrubEnd = adjustOffsetForZone(
      $dashboardStore?.selectedScrubRange?.end,
      $dashboardStore?.selectedTimezone
    );
  }

  let mouseoverValue = undefined;
  let startValue: Date;
  let endValue: Date;

  // FIXME: move this logic to a function + write tests.
  $: if ($timeControlsStore.ready) {
    const adjustedChartValue = getAdjustedChartTime(
      $timeControlsStore.selectedTimeRange?.start,
      $timeControlsStore.selectedTimeRange?.end,
      $dashboardStore?.selectedTimezone,
      interval,
      $timeControlsStore.selectedTimeRange?.name
    );

    startValue = adjustedChartValue?.start;
    endValue = adjustedChartValue?.end;
  }

  $: showHideMeasures = createShowHideMeasuresStore(metricViewName, metaQuery);

  const toggleMeasureVisibility = (e) => {
    showHideMeasures.toggleVisibility(e.detail.name);
  };
  const setAllMeasuresNotVisible = () => {
    showHideMeasures.setAllToNotVisible();
  };
  const setAllMeasuresVisible = () => {
    showHideMeasures.setAllToVisible();
  };

  function onKeyDown(e) {
    if (scrubStart && scrubEnd) {
      // if key Z is pressed, zoom the scrub
      if (e.key === "z") {
        zoomScrub();
      } else if (
        !$dashboardStore.selectedScrubRange?.isScrubbing &&
        e.key === "Escape"
      ) {
        metricsExplorerStore.setSelectedScrubRange(metricViewName, undefined);
      }
    }
  }

  function zoomScrub() {
    const { start, end } = getOrderedStartEnd(
      $dashboardStore?.selectedScrubRange?.start,
      $dashboardStore?.selectedScrubRange?.end
    );
    metricsExplorerStore.setSelectedTimeRange(metricViewName, {
      name: TimeRangePreset.CUSTOM,
      start,
      end,
    });
  }
</script>

<TimeSeriesChartContainer end={endValue} start={startValue} {workspaceWidth}>
  <div class="bg-white sticky top-0 flex" style="z-index:100">
    <SeachableFilterButton
      label="Measures"
      on:deselect-all={setAllMeasuresNotVisible}
      on:item-clicked={toggleMeasureVisibility}
      on:select-all={setAllMeasuresVisible}
      selectableItems={$showHideMeasures.selectableItems}
      selectedItems={$showHideMeasures.selectedItems}
      tooltipText="Choose measures to display"
    />
  </div>
  <div
    class="bg-white sticky left-0 top-0 overflow-visible"
    style="z-index:101"
  >
    <!-- top axis element -->
    <div />
    <div bind:this={axisTop} style:height="24px" style:padding-left="24px">
      {#if $dashboardStore?.selectedScrubRange?.end && !$dashboardStore?.selectedScrubRange?.isScrubbing}
        <Portal>
          <FloatingElement
            target={axisTop}
            location="top"
            relationship="direct"
            alignment="middle"
            distance={10}
            pad={0}
          >
            <div style:left="-40px" class="absolute flex justify-center">
              <Button compact type="highlighted" on:click={() => zoomScrub()}>
                <div class="flex items-center gap-x-2">
                  <Zoom size="16px" />
                  Zoom
                  <span class="font-semibold">(Z)</span>
                </div>
              </Button>
            </div>
          </FloatingElement>
        </Portal>
      {/if}
    </div>
    {#if $dashboardStore?.selectedTimeRange}
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
  {#if $metaQuery.data?.measures}
    <!-- FIXME: this is pending the remaining state work for show/hide measures and dimensions -->
    {#each $metaQuery.data?.measures.filter((_, i) => $showHideMeasures.selectedItems[i]) as measure, index (measure.name)}
      <!-- FIXME: I can't select the big number by the measure id. -->
      {@const bigNum = $totalsQuery?.data?.data?.[measure.name]}
      {@const comparisonValue = totalsComparisons?.[measure.name]}
      {@const comparisonPercChange =
        comparisonValue && bigNum !== undefined && bigNum !== null
          ? (bigNum - comparisonValue) / comparisonValue
          : undefined}
      {@const formatPreset =
        FormatPreset[measure?.format] || FormatPreset.HUMANIZE}
      <!-- FIXME: I can't select a time series by measure id. -->
      <MeasureBigNumber
        value={bigNum}
        {showComparison}
        comparisonOption={$timeControlsStore?.selectedComparisonTimeRange?.name}
        {comparisonValue}
        {comparisonPercChange}
        description={measure?.description ||
          measure?.label ||
          measure?.expression}
        formatPreset={measure?.format}
        status={$totalsQuery?.isFetching
          ? EntityStatus.Running
          : EntityStatus.Idle}
      >
        <svelte:fragment slot="name">
          {measure?.label || measure?.expression}
        </svelte:fragment>
      </MeasureBigNumber>
      <div class="time-series-body" style:height="125px">
        {#if $timeSeriesQuery?.isError}
          <div class="p-5"><CrossIcon /></div>
        {:else if formattedData}
          <MeasureChart
            isScrubbing={$dashboardStore?.selectedScrubRange?.isScrubbing}
            {scrubStart}
            {scrubEnd}
            bind:mouseoverValue
            {metricViewName}
            data={formattedData}
            zone={$dashboardStore?.selectedTimezone}
            xAccessor="ts_position"
            labelAccessor="ts"
            timeGrain={interval}
            yAccessor={measure.name}
            xMin={startValue}
            xMax={endValue}
            {showComparison}
            mouseoverTimeFormat={(value) => {
              /** format the date according to the time grain */
              return new Date(value).toLocaleDateString(
                undefined,
                TIME_GRAIN[interval].formatDate
              );
            }}
            numberKind={nicelyFormattedTypesToNumberKind(measure?.format)}
            mouseoverFormat={(value) =>
              formatPreset === FormatPreset.NONE
                ? `${value}`
                : humanizeDataType(value, measure?.format)}
          />
        {:else}
          <div>
            <Spinner status={EntityStatus.Running} />
          </div>
        {/if}
      </div>
    {/each}
  {/if}
</TimeSeriesChartContainer>

<!-- Only to be used on singleton components to avoid multiple state dispatches -->
<svelte:window on:keydown={onKeyDown} />
