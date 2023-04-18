<script lang="ts">
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import {
    humanizeDataType,
    NicelyFormattedTypes,
    nicelyFormattedTypesToNumberKind,
  } from "@rilldata/web-common/features/dashboards/humanize-numbers";
  import {
    useMetaQuery,
    useModelAllTimeRange,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { removeTimezoneOffset } from "@rilldata/web-common/lib/formatters";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import {
    getDurationMultiple,
    getEndOfPeriod,
    getOffset,
    getStartOfPeriod,
  } from "@rilldata/web-common/lib/time/transforms";
  import { TimeOffsetType } from "@rilldata/web-common/lib/time/types";
  import {
    createQueryServiceMetricsViewTimeSeries,
    createQueryServiceMetricsViewTotals,
    V1MetricsViewTimeSeriesResponse,
    V1MetricsViewTotalsResponse,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import type { CreateQueryResult } from "@tanstack/svelte-query";
  import { isRangeInsideOther } from "../../../lib/time/ranges";
  import { runtime } from "../../../runtime-client/runtime-store";
  import Spinner from "../../entity-management/Spinner.svelte";
  import MeasureBigNumber from "../big-number/MeasureBigNumber.svelte";
  import MeasureChart from "./MeasureChart.svelte";
  import TimeSeriesChartContainer from "./TimeSeriesChartContainer.svelte";
  import { prepareTimeSeries } from "./utils";

  export let metricViewName;
  export let workspaceWidth: number;

  $: dashboardStore = useDashboardStore(metricViewName);

  $: instanceId = $runtime.instanceId;

  // query the `/meta` endpoint to get the measures and the default time grain
  $: metaQuery = useMetaQuery(instanceId, metricViewName);
  $: selectedMeasureNames = $dashboardStore?.selectedMeasureNames;
  $: interval = $dashboardStore?.selectedTimeRange?.interval;

  let totalsQuery: CreateQueryResult<V1MetricsViewTotalsResponse, Error>;

  $: allTimeRangeQuery = useModelAllTimeRange(
    $runtime.instanceId,
    $metaQuery.data.model,
    $metaQuery.data.timeDimension
  );

  // get the time range name, which is the preset.
  let name;
  let allTimeRange;
  $: if ($allTimeRangeQuery?.isSuccess) {
    allTimeRange = $allTimeRangeQuery.data;
    name = $dashboardStore.selectedTimeRange.name;
  }

  let totalsComparisonQuery: CreateQueryResult<
    V1MetricsViewTotalsResponse,
    Error
  >;

  /** Get extra data point for interpolating the chart for the first point*/
  function getAdjustedStartTime(date: Date, interval: V1TimeGrain) {
    if (!date) return undefined;
    const offsetedDate = getOffset(
      date,
      TIME_GRAIN[interval].duration,
      TimeOffsetType.SUBTRACT
    );

    // the data point previous to the first date inside the chart.
    const trucatedOffsetedDate = getStartOfPeriod(
      offsetedDate,
      TIME_GRAIN[interval].duration
    );

    return trucatedOffsetedDate.toISOString();
  }

  let isComparisonRangeAvailable = false;

  /** Generate the totals & big number comparison query */
  $: if (
    name &&
    $dashboardStore &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching &&
    allTimeRange?.start &&
    $dashboardStore?.selectedTimeRange?.start
  ) {
    isComparisonRangeAvailable = isRangeInsideOther(
      allTimeRange?.start,
      allTimeRange?.end,
      $dashboardStore?.selectedComparisonTimeRange?.start,
      $dashboardStore?.selectedComparisonTimeRange?.end
    );

    const totalsQueryParams = {
      measureNames: selectedMeasureNames,
      filter: $dashboardStore?.filters,
      timeStart: $dashboardStore.selectedTimeRange?.start.toISOString(),
      timeEnd: $dashboardStore.selectedTimeRange?.end.toISOString(),
    };

    totalsQuery = createQueryServiceMetricsViewTotals(
      instanceId,
      metricViewName,
      totalsQueryParams
    );

    totalsComparisonQuery = createQueryServiceMetricsViewTotals(
      instanceId,
      metricViewName,
      {
        ...totalsQueryParams,
        timeStart: isComparisonRangeAvailable
          ? $dashboardStore?.selectedComparisonTimeRange?.start.toISOString()
          : undefined,
        timeEnd: isComparisonRangeAvailable
          ? $dashboardStore?.selectedComparisonTimeRange?.end.toISOString()
          : undefined,
      }
    );
  }

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
    $dashboardStore?.selectedTimeRange?.start
  ) {
    timeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
      instanceId,
      metricViewName,
      {
        measureNames: selectedMeasureNames,
        filter: $dashboardStore?.filters,
        timeStart: getAdjustedStartTime(
          $dashboardStore.selectedTimeRange?.start,
          interval
        ),
        timeEnd: $dashboardStore.selectedTimeRange?.end.toISOString(),
        timeGranularity: interval,
      }
    );
    if (isComparisonRangeAvailable) {
      timeSeriesComparisonQuery = createQueryServiceMetricsViewTimeSeries(
        instanceId,
        metricViewName,
        {
          measureNames: selectedMeasureNames,
          filter: $dashboardStore?.filters,
          timeStart: getAdjustedStartTime(
            $dashboardStore?.selectedComparisonTimeRange?.start,
            interval
          ),
          timeEnd:
            $dashboardStore?.selectedComparisonTimeRange?.end.toISOString(),
          timeGranularity: interval,
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

  $: if ($timeSeriesQuery?.data?.data) dataCopy = $timeSeriesQuery.data.data;
  $: if ($timeSeriesComparisonQuery?.data?.data)
    dataComparisonCopy = $timeSeriesComparisonQuery.data.data;

  // formattedData adjusts the data to account for Javascript's handling of timezones
  let formattedData;
  $: if (dataCopy && dataCopy?.length) {
    formattedData = prepareTimeSeries(
      dataCopy,
      dataComparisonCopy,
      TIME_GRAIN[interval].duration
    );
  }

  let mouseoverValue = undefined;

  let startValue: Date;
  let endValue: Date;

  // FIXME: move this logic to a function + write tests.
  $: if (
    $dashboardStore?.selectedTimeRange &&
    $dashboardStore?.selectedTimeRange?.start
  ) {
    startValue = removeTimezoneOffset(
      new Date($dashboardStore?.selectedTimeRange?.start)
    );

    // selectedTimeRange.end is exclusive and rounded to the time grain ("interval").
    endValue = new Date($dashboardStore?.selectedTimeRange?.end);
    endValue = removeTimezoneOffset(endValue);
  }
</script>

<TimeSeriesChartContainer {workspaceWidth} start={startValue} end={endValue}>
  <div class="bg-white sticky left-0 top-0" />
  <div class="bg-white sticky left-0 top-0">
    <div style:padding-left="24px" style:height="20px" />
    <!-- top axis element -->
    <div />
    {#if $dashboardStore?.selectedTimeRange}
      <SimpleDataGraphic
        height={32}
        top={34}
        bottom={0}
        xMin={startValue}
        xMax={endValue}
      >
        <Axis superlabel side="top" />
      </SimpleDataGraphic>
    {/if}
  </div>
  <!-- bignumbers and line charts -->
  {#if $metaQuery.data?.measures && $totalsQuery?.isSuccess}
    {#each $metaQuery.data?.measures as measure, index (measure.name)}
      <!-- FIXME: I can't select the big number by the measure id. -->
      {@const bigNum = $totalsQuery?.data.data?.[measure.name]}
      {@const showComparison = isComparisonRangeAvailable}
      {@const comparisonValue = totalsComparisons?.[measure.name]}
      {@const comparisonPercChange =
        comparisonValue && bigNum !== undefined && bigNum !== null
          ? (bigNum - comparisonValue) / comparisonValue
          : undefined}
      {@const formatPreset =
        NicelyFormattedTypes[measure?.format] || NicelyFormattedTypes.HUMANIZE}
      <!-- FIXME: I can't select a time series by measure id. -->
      <MeasureBigNumber
        value={bigNum}
        {showComparison}
        comparisonOption={$dashboardStore?.selectedComparisonTimeRange?.name}
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
            data={formattedData}
            xAccessor="ts_position"
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
              formatPreset === NicelyFormattedTypes.NONE
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
