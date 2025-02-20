<script lang="ts">
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import {
    createQueryServiceMetricsViewAggregation,
    createQueryServiceMetricsViewTimeSeries,
    V1TimeGrain,
    type V1ComponentSpecRendererProperties,
  } from "@rilldata/web-common/runtime-client";
  import type { KPISpec } from ".";
  import { validateKPISchema } from "./selector";
  import Chart from "@rilldata/web-common/components/time-series-chart/Chart.svelte";
  import { Interval } from "luxon";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let rendererProperties: V1ComponentSpecRendererProperties;

  const ctx = getCanvasStateManagers();
  const {
    spec,
    timeControls: { showTimeComparison, selectedComparisonTimeRange },
  } = ctx.canvasEntity;

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
    comparison_range: comparisonTimeRange,
  } = kpiProperties);

  $: schema = validateKPISchema(ctx, kpiProperties);

  $: measureStore = spec.getMeasureForMetricView(measureName, metricsViewName);

  $: measure = $measureStore;
  $: measureIsPercentage = measure?.formatPreset === FormatPreset.PERCENTAGE;

  $: measureValueFormatter = measure
    ? createMeasureValueFormatter<null>(measure, "big-number")
    : () => "no data";

  $: showSparkline = sparkline !== "none";
  $: isSparkRight = sparkline === "right";

  $: ({ isValid } = $schema);

  $: timeAndFilterStore = ctx.canvasEntity.createTimeAndFilterStore(
    metricsViewName,
    {
      componentFilter: kpiProperties.dimension_filters,
      componentComparisonRange: comparisonTimeRange,
    },
  );

  $: ({
    timeGrain,
    timeRange: { timeZone, start, end },
    where,
    comparisonRange: { start: comparisonStart, end: comparisonEnd },
  } = $timeAndFilterStore);

  $: showComparison =
    ($showTimeComparison || !!comparisonTimeRange) && !!comparisonOptions;

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
      timeRange: {
        start: comparisonStart,
        end: comparisonEnd,
        timeZone,
      },
      where,
    },
    {
      query: {
        enabled: showComparison && isValid && !!start && !!end,
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
    currentValue != null && comparisonVal != null
      ? (currentValue - comparisonVal) / comparisonVal
      : undefined;

  $: globalComparisonLabel =
    $selectedComparisonTimeRange?.name &&
    TIME_COMPARISON[$selectedComparisonTimeRange?.name]?.label;

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
      timeStart: comparisonStart,
      timeEnd: comparisonEnd,
      timeGranularity: timeGrain || V1TimeGrain.TIME_GRAIN_HOUR,
      timeZone,
      where,
    },
    {
      query: {
        enabled:
          !!comparisonStart &&
          !!comparisonEnd &&
          isValid &&
          showSparkline &&
          showComparison,
      },
    },
  );

  $: ({ data: sparkDataResponse, isFetching: primaryDataIsFetching } =
    $sparklineDataQuery);
  $: primaryData = sparkDataResponse?.data ?? [];

  $: ({ data: comparisonDataResponse } = $comparisonDataQuery);
  $: comparisonData = comparisonDataResponse?.data ?? [];

  function getFormattedDiff(comparisonValue: number, currentValue: number) {
    const delta = currentValue - comparisonValue;
    return `${delta >= 0 ? "+" : ""}${measureValueFormatter(delta)}`;
  }
</script>

{#if isValid}
  {#if measure && !primaryTotalIsFetching}
    <div
      class:flex-col={!isSparkRight}
      class="flex gap-0 items-center justify-center size-full"
    >
      <div class:!items-start={isSparkRight} class="flex flex-col items-center">
        <h2
          class:text-center={!isSparkRight}
          class="font-medium text-sm text-gray-600"
        >
          {measure?.displayName || measureName}
        </h2>
        <span class="text-3xl font-medium text-gray-600">
          {measureValueFormatter(currentValue)}
        </span>
      </div>

      <div class="flex flex-col items-center">
        {#if showComparison && comparisonVal != null}
          <div class="flex items-baseline gap-x-2 text-sm -mb-[3px]">
            {#if comparisonOptions?.includes("previous")}
              <span class="comparison-value">
                {measureValueFormatter(comparisonVal)}
              </span>
            {/if}

            {#if comparisonOptions?.includes("delta")}
              <span
                class="comparison-value"
                class:text-red-500={primaryTotal &&
                  primaryTotal - comparisonVal < 0}
                class:ui-copy-disabled-faint={comparisonVal === null}
                class:italic={comparisonVal === null}
                class:text-sm={comparisonVal === null}
              >
                {#if comparisonVal != null}
                  {getFormattedDiff(comparisonVal, currentValue)}
                {:else}
                  no change
                {/if}
              </span>
            {/if}

            {#if comparisonOptions?.includes("percent_change") && comparisonPercChange != null && !measureIsPercentage}
              <div
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
              </div>
            {/if}
          </div>

          {#if comparisonTimeRange || globalComparisonLabel}
            <div class="comparison-range">
              vs {comparisonTimeRange
                ? `last ${humaniseISODuration(comparisonTimeRange?.toUpperCase(), false)}`
                : globalComparisonLabel?.toLowerCase()}
            </div>
          {/if}
        {/if}
      </div>

      <div class="size-full flex items-center justify-center mt-2">
        {#if primaryDataIsFetching}
          <Spinner status={EntityStatus.Running} />
        {:else if timeGrain && timeZone}
          <Chart
            bind:hoveredPoints
            {primaryData}
            secondaryData={showComparison ? [comparisonData] : []}
            {timeGrain}
            selectedTimeZone={timeZone}
            yAccessor={kpiProperties.measure}
            formatterFunction={measureValueFormatter}
          />
        {/if}
      </div>
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
  .measure-label {
    @apply font-medium text-sm whitespace-normal;
    @apply pr-2 text-gray-700;
  }
  .measure-value {
    @apply text-3xl font-medium text-gray-700 pb-1;
  }
  .comparison-range {
    @apply text-sm text-gray-400;
  }

  .comparison-value {
    @apply w-fit max-w-full overflow-hidden;
    @apply font-medium text-ellipsis text-gray-500;
  }
</style>
