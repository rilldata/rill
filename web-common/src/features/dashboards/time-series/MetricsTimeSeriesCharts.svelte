<script lang="ts">
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import SeachableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SeachableFilterButton.svelte";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import { getFilterForComparedDimension, prepareTimeSeries } from "./utils";
  import {
    humanizeDataType,
    FormatPreset,
    nicelyFormattedTypesToNumberKind,
  } from "@rilldata/web-common/features/dashboards/humanize-numbers";
  import {
    getFilterForDimension,
    useMetaQuery,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { createShowHideMeasuresStore } from "@rilldata/web-common/features/dashboards/show-hide-selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
  import { getAdjustedChartTime } from "@rilldata/web-common/lib/time/ranges";
  import {
    createQueryServiceMetricsViewTimeSeries,
    createQueryServiceMetricsViewToplist,
    createQueryServiceMetricsViewTotals,
    V1MetricsViewTimeSeriesResponse,
  } from "@rilldata/web-common/runtime-client";
  import type { CreateQueryResult } from "@tanstack/svelte-query";
  import { getDimensionValueTimeSeries } from "./multiple-dimension-queries";
  import { runtime } from "../../../runtime-client/runtime-store";
  import Spinner from "../../entity-management/Spinner.svelte";
  import MeasureBigNumber from "../big-number/MeasureBigNumber.svelte";
  import MeasureChart from "./MeasureChart.svelte";
  import MeasureZoom from "./MeasureZoom.svelte";
  import TimeSeriesChartContainer from "./TimeSeriesChartContainer.svelte";

  export let metricViewName;
  export let workspaceWidth: number;

  $: dashboardStore = useDashboardStore(metricViewName);

  $: instanceId = $runtime.instanceId;

  // query the `/meta` endpoint to get the measures and the default time grain
  $: metaQuery = useMetaQuery(instanceId, metricViewName);

  const timeControlsStore = useTimeControlStore(getStateManagers());

  $: selectedMeasureNames = $dashboardStore?.selectedMeasureNames;
  $: comparisonDimension = $dashboardStore?.selectedComparisonDimension;
  $: showComparison = !comparisonDimension && $timeControlsStore.showComparison;
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

  let includedValues;
  let allDimQuery;

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

  let topListQuery;
  $: if (comparisonDimension && $timeControlsStore.ready) {
    const dimensionFilters = $dashboardStore.filters.include.filter(
      (filter) => filter.name === comparisonDimension
    );
    if (dimensionFilters) {
      includedValues = dimensionFilters[0]?.in.slice(0, 7) || [];
    }

    if (includedValues.length === 0) {
      // TODO: Create a central store for topList
      // Fetch top values for the dimension
      const filterForDimension = getFilterForDimension(
        $dashboardStore?.filters,
        comparisonDimension
      );
      topListQuery = createQueryServiceMetricsViewToplist(
        $runtime.instanceId,
        metricViewName,
        {
          dimensionName: comparisonDimension,
          measureNames: [$dashboardStore?.leaderboardMeasureName],
          timeStart: $timeControlsStore.timeStart,
          timeEnd: $timeControlsStore.timeEnd,
          filter: filterForDimension,
          limit: "250",
          offset: "0",
          sort: [
            {
              name: $dashboardStore?.leaderboardMeasureName,
              ascending:
                $dashboardStore.sortDirection === SortDirection.ASCENDING,
            },
          ],
        },
        {
          query: {
            enabled: $timeControlsStore.ready && !!filterForDimension,
          },
        }
      );
    }
  }

  $: if (
    includedValues?.length ||
    (topListQuery && !$topListQuery?.isFetching)
  ) {
    let filters = $dashboardStore.filters;

    // Handle case when there are no included filters for the dimension
    if (!includedValues?.length) {
      const columnName = $topListQuery?.data?.meta[0]?.name;
      const topListValues = $topListQuery?.data?.data.map((d) => d[columnName]);

      const computedFilter = getFilterForComparedDimension(
        comparisonDimension,
        $dashboardStore?.filters,
        topListValues
      );
      filters = computedFilter?.updatedFilter;
      includedValues = computedFilter?.includedValues;
    }

    allDimQuery = getDimensionValueTimeSeries(
      includedValues,
      instanceId,
      metricViewName,
      comparisonDimension,
      selectedMeasureNames,
      filters,
      $timeControlsStore.adjustedStart,
      $timeControlsStore.adjustedEnd,
      interval,
      $dashboardStore?.selectedTimezone
    );
  }

  $: dimensionData = comparisonDimension ? $allDimQuery : [];

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
</script>

<TimeSeriesChartContainer end={endValue} start={startValue} {workspaceWidth}>
  <div class="bg-white sticky top-0 flex pl-1" style="z-index:100">
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
    <MeasureZoom {metricViewName} />
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
    {#each $metaQuery.data?.measures.filter((_, i) => $showHideMeasures.selectedItems[i]) as measure (measure.name)}
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
            bind:mouseoverValue
            isScrubbing={$dashboardStore?.selectedScrubRange?.isScrubbing}
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
