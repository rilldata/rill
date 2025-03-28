<script lang="ts">
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import Chart from "@rilldata/web-common/components/time-series-chart/Chart.svelte";
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
  import RangeDisplay from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/RangeDisplay.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import {
    createQueryServiceMetricsViewAggregation,
    createQueryServiceMetricsViewTimeSeries,
    V1TimeGrain,
    type V1ComponentSpecRendererProperties,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { DateTime, Interval } from "luxon";
  import type { Readable } from "svelte/motion";
  import type { KPISpec } from ".";
  import { BIG_NUMBER_MIN_WIDTH } from ".";
  import { validateKPISchema } from "./selector";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;

  const ctx = getCanvasStateManagers();
  const { spec } = ctx.canvasEntity;

  let hoveredPoints: {
    interval: Interval<true>;
    value: number | null | undefined;
  }[] = [];

  $: ({ instanceId } = $runtime);

  $: kpiProperties = rendererProperties as KPISpec;

  $: ({
    metrics_view: metricsViewName,
    measure: measureName,
    sparkline,
    comparison: comparisonOptions,
  } = kpiProperties);

  $: ({
    timeGrain,
    timeRange: { timeZone, start, end },
    where,
    comparisonTimeRange,
    showTimeComparison,
    comparisonTimeRangeState,
    hasTimeSeries,
  } = $timeAndFilterStore);

  $: schema = validateKPISchema(ctx, kpiProperties);
  $: ({ isValid } = $schema);

  $: measureStore = spec.getMeasureForMetricView(measureName, metricsViewName);
  $: measure = $measureStore;
  $: measureIsPercentage = measure?.formatPreset === FormatPreset.PERCENTAGE;

  $: measureValueFormatter = measure
    ? createMeasureValueFormatter<null>(measure, "big-number")
    : () => "no data";

  $: showSparkline = sparkline !== "none" && hasTimeSeries;
  $: isSparkRight = sparkline === "right";

  $: showComparison = !!comparisonOptions?.length && showTimeComparison;

  $: comparisonLabel =
    comparisonTimeRangeState?.selectedComparisonTimeRange?.name &&
    TIME_COMPARISON[comparisonTimeRangeState?.selectedComparisonTimeRange.name]
      ?.label;

  // BIG NUMBER QUERIES
  $: kpiTotalsQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: [{ name: measureName }],
      timeRange: {
        start,
        end,
        timeZone,
      },
      where,
    },
    {
      query: {
        enabled: isValid && !!start && !!end,
      },
    },
  );

  $: kpiComparisonTotalsQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: [{ name: measureName }],
      timeRange: comparisonTimeRange,
      where,
    },
    {
      query: {
        enabled:
          comparisonTimeRange && showComparison && isValid && !!start && !!end,
      },
    },
  );

  $: ({ data: primaryTotalData, isFetching: primaryTotalIsFetching } =
    $kpiTotalsQuery);
  $: ({ data: comparisonTotalData } = $kpiComparisonTotalsQuery);

  $: primaryTotal = (primaryTotalData?.data?.[0]?.[measureName] ?? null) as
    | number
    | null;
  $: comparisonTotal = (comparisonTotalData?.data?.[0]?.[measureName] ??
    null) as number | null;

  $: currentValue =
    hoveredPoints?.[0]?.value !== undefined
      ? hoveredPoints?.[0]?.value
      : primaryTotal;

  $: comparisonVal =
    hoveredPoints?.[1]?.value !== undefined
      ? hoveredPoints?.[1]?.value
      : comparisonTotal;

  $: comparisonPercChange =
    currentValue != null && comparisonVal
      ? (currentValue - comparisonVal) / comparisonVal
      : null;

  // TIME SERIES QUERIES
  $: sparklineDataQuery = createQueryServiceMetricsViewTimeSeries(
    instanceId,
    metricsViewName,
    {
      measureNames: [measureName],
      timeStart: start,
      timeEnd: end,
      timeGranularity: timeGrain || V1TimeGrain.TIME_GRAIN_HOUR,
      timeZone,
      where,
    },
    {
      query: {
        enabled: !!start && !!end && $schema.isValid && showSparkline,
      },
    },
  );

  $: comparisonDataQuery = createQueryServiceMetricsViewTimeSeries(
    instanceId,
    metricsViewName,
    {
      measureNames: [measureName],
      timeStart: comparisonTimeRange?.start,
      timeEnd: comparisonTimeRange?.end,
      timeGranularity: timeGrain || V1TimeGrain.TIME_GRAIN_HOUR,
      timeZone,
      where,
    },
    {
      query: {
        enabled:
          comparisonTimeRange && isValid && showSparkline && showComparison,
      },
    },
  );

  $: ({ data: sparkDataResponse, isFetching: primaryDataIsFetching } =
    $sparklineDataQuery);
  $: primaryData = sparkDataResponse?.data ?? [];

  $: ({ data: comparisonDataResponse } = $comparisonDataQuery);
  $: comparisonData = comparisonDataResponse?.data ?? [];

  $: interval = Interval.fromDateTimes(
    DateTime.fromISO(start ?? "").setZone(timeZone),
    DateTime.fromISO(end ?? "").setZone(timeZone),
  );

  $: adjustment = 30 - (comparisonOptions?.length ?? 0) * 10;

  function getFormattedDiff(comparisonValue: number, currentValue: number) {
    const delta = currentValue - comparisonValue;
    return `${delta >= 0 ? "+" : ""}${measureValueFormatter(delta)}`;
  }
</script>

{#if isValid}
  {#if measure && primaryTotalData && !primaryTotalIsFetching}
    <div class="wrapper" class:spark-right={isSparkRight}>
      <div
        class="data-wrapper"
        style:min-width="{BIG_NUMBER_MIN_WIDTH - adjustment}px"
      >
        <h2 class="measure-name" title={measure?.displayName || measureName}>
          {measure?.displayName || measureName}
        </h2>

        <span
          class="big-number"
          class:hovered-value={hoveredPoints?.[0]?.value != null}
        >
          {measureValueFormatter(
            hoveredPoints?.[0]?.value != null ? currentValue : primaryTotal,
          )}
        </span>

        {#if showComparison}
          <div class="comparison-value-wrapper">
            {#if comparisonOptions?.includes("previous")}
              <span class="comparison-value">
                {measureValueFormatter(comparisonVal)}
              </span>
            {/if}

            {#if comparisonOptions?.includes("delta")}
              <span
                class="comparison-value"
                class:text-red-500={primaryTotal !== null &&
                  comparisonVal !== null &&
                  primaryTotal - comparisonVal < 0}
                class:ui-copy-disabled-faint={comparisonVal === null}
                class:italic={comparisonVal === null}
                class:text-sm={comparisonVal === null}
              >
                {#if comparisonVal != null && currentValue != null}
                  {getFormattedDiff(comparisonVal, currentValue)}
                {:else}
                  no change
                {/if}
              </span>
            {/if}

            {#if comparisonOptions?.includes("percent_change") && comparisonPercChange != null && !measureIsPercentage}
              <span
                class="w-fit font-semibold ui-copy-inactive"
                class:text-red-500={primaryTotal && primaryTotal < 0}
              >
                <PercentageChange
                  color="text-gray-500"
                  showPosSign
                  tabularNumber={false}
                  value={formatMeasurePercentageDifference(
                    comparisonPercChange,
                  )}
                />
              </span>
            {/if}
          </div>

          {#if comparisonLabel}
            <p class="text-sm text-gray-400 break-words">
              vs {comparisonLabel?.toLowerCase()}
            </p>
          {/if}
        {/if}

        {#if !showSparkline && timeGrain && interval.isValid}
          <span class="text-gray-500">
            <RangeDisplay {interval} grain={timeGrain} />
          </span>
        {/if}
      </div>

      {#if showSparkline}
        <div class="sparkline-wrapper">
          {#if primaryDataIsFetching}
            <Spinner status={EntityStatus.Running} />
          {:else if timeGrain && timeZone && primaryData.length}
            <Chart
              bind:hoveredPoints
              {primaryData}
              secondaryData={showComparison ? comparisonData : []}
              {timeGrain}
              selectedTimeZone={timeZone}
              yAccessor={kpiProperties.measure}
              formatterFunction={measureValueFormatter}
            />
          {/if}
        </div>
      {/if}
    </div>
  {:else}
    <div class="flex items-center justify-center w-full h-full">
      <Spinner size="36px" status={EntityStatus.Running} />
    </div>
  {/if}
{:else}
  <ComponentError error={$schema.error} />
{/if}

<style lang="postcss">
  .wrapper {
    @apply flex items-center justify-center size-full gap-2 flex-col max-w-full;
  }

  .wrapper.spark-right {
    @apply flex-row;
  }

  .data-wrapper {
    @apply flex flex-col w-full h-fit justify-center items-center max-w-full;
    @apply overflow-hidden text-ellipsis truncate;
    flex: 1 0 auto;
  }

  .spark-right .data-wrapper {
    @apply items-start h-full;
    flex: 0 4 20%;
  }

  .measure-name {
    @apply flex-none text-center font-medium text-sm text-gray-600 break-words line-clamp-1;
  }

  .spark-right .measure-name {
    @apply line-clamp-2 text-left max-w-40;
  }

  .big-number {
    @apply text-3xl font-medium text-gray-600 flex gap-x-0.5 items-end;
  }

  .hovered-value {
    @apply text-primary-500;
  }

  .comparison-value-wrapper {
    @apply flex items-baseline gap-x-2 text-sm -mb-[3px] truncate flex-none;
  }

  .sparkline-wrapper {
    @apply size-full flex items-center justify-center flex-shrink;
  }

  .spark-right .sparkline-wrapper {
    @apply mt-2;
    flex: 4 1 80%;
  }

  .comparison-value {
    @apply w-fit max-w-full overflow-hidden;
    @apply font-medium text-ellipsis text-gray-500;
  }

  @container component-container (inline-size < 300px) {
    .spark-right .sparkline-wrapper {
      display: none;
    }

    .spark-right .data-wrapper {
      align-items: center !important;
      flex: 0 2 auto !important;
    }

    .spark-right .measure-name {
      max-width: 100% !important;
      text-align: center !important;
    }
  }
</style>
