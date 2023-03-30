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
  import { getOffset } from "@rilldata/web-common/lib/time/transforms";
  import { TimeOffsetType } from "@rilldata/web-common/lib/time/types";
  import {
    useQueryServiceMetricsViewTimeSeries,
    useQueryServiceMetricsViewTotals,
    V1MetricsViewTimeSeriesResponse,
    V1MetricsViewTotalsResponse,
  } from "@rilldata/web-common/runtime-client";
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
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
  $: timeDimension = $metaQuery.data?.timeDimension;
  $: selectedMeasureNames = $dashboardStore?.selectedMeasureNames;
  $: interval = $dashboardStore?.selectedTimeRange?.interval;

  let totalsQuery: UseQueryStoreResult<V1MetricsViewTotalsResponse, Error>;

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

  let totalsComparisonQuery: UseQueryStoreResult<
    V1MetricsViewTotalsResponse,
    Error
  >;

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

    totalsQuery = useQueryServiceMetricsViewTotals(
      instanceId,
      metricViewName,
      totalsQueryParams
    );

    totalsComparisonQuery = useQueryServiceMetricsViewTotals(
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

  let timeSeriesQuery: UseQueryStoreResult<
    V1MetricsViewTimeSeriesResponse,
    Error
  >;

  let timeSeriesComparisonQuery: UseQueryStoreResult<
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
    timeSeriesQuery = useQueryServiceMetricsViewTimeSeries(
      instanceId,
      metricViewName,
      {
        measureNames: selectedMeasureNames,
        filter: $dashboardStore?.filters,
        timeStart: $dashboardStore.selectedTimeRange?.start.toISOString(),
        timeEnd: $dashboardStore.selectedTimeRange?.end.toISOString(),
        timeGranularity: $dashboardStore.selectedTimeRange?.interval,
      }
    );
    if (isComparisonRangeAvailable) {
      timeSeriesComparisonQuery = useQueryServiceMetricsViewTimeSeries(
        instanceId,
        metricViewName,
        {
          measureNames: selectedMeasureNames,
          filter: $dashboardStore?.filters,
          timeStart:
            $dashboardStore?.selectedComparisonTimeRange?.start.toISOString(),
          timeEnd:
            $dashboardStore?.selectedComparisonTimeRange?.end.toISOString(),
          timeGranularity: $dashboardStore.selectedTimeRange?.interval,
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
    formattedData = prepareTimeSeries(dataCopy, dataComparisonCopy);
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
    // Since values are grouped with DATE_TRUNC, we subtract one grain to get the (inclusive) axis end.
    endValue = new Date($dashboardStore?.selectedTimeRange?.end);

    endValue = getOffset(
      new Date($dashboardStore?.selectedTimeRange?.end),
      TIME_GRAIN[$dashboardStore?.selectedTimeRange?.interval].duration,
      TimeOffsetType.SUBTRACT
    );

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
        comparisonValue && bigNum
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
            xAccessor="ts"
            timeGrain={$dashboardStore?.selectedTimeRange?.interval}
            yAccessor={measure.name}
            xMin={startValue}
            xMax={endValue}
            start={startValue}
            end={endValue}
            {showComparison}
            mouseoverTimeFormat={(value) => {
              /** format the date according to the time grain */
              return new Date(value).toLocaleDateString(
                undefined,
                TIME_GRAIN[$dashboardStore?.selectedTimeRange?.interval]
                  .formatDate
              );
            }}
            numberKind={nicelyFormattedTypesToNumberKind(measure?.format)}
            mouseoverFormat={(value) =>
              formatPreset === NicelyFormattedTypes.NONE
                ? `${value}`
                : humanizeDataType(value, measure?.format, {
                    excludeDecimalZeros: true,
                  })}
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
