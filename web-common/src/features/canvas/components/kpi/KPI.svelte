<script lang="ts">
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import {
    FormatPreset,
    NumberKind,
    numberKindForMeasure,
  } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import {
    createQueryServiceMetricsViewTimeSeries,
    V1TimeGrain,
    type V1ComponentSpecRendererProperties,
  } from "@rilldata/web-common/runtime-client";
  import type { KPISpec } from ".";
  import KPISparkline from "./KPISparkline.svelte";
  import {
    useKPIComparisonTotal,
    useKPISparkline,
    useKPITotals,
    validateKPISchema,
  } from "./selector";
  import Chart from "@rilldata/web-common/components/time-series-chart/Chart.svelte";
  import { DateTime, Interval } from "luxon";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let topPadding = true;

  const ctx = getCanvasStateManagers();
  const {
    spec,
    timeControls: {
      showTimeComparison,
      selectedComparisonTimeRange,
      selectedTimeRange,
    },
  } = ctx.canvasEntity;

  let containerWidth: number;
  let containerHeight: number;

  $: kpiProperties = rendererProperties as KPISpec;

  $: ({
    metrics_view: metricsViewName,
    measure: measureName,
    sparkline,
    comparison: comparisonOptions,
    comparison_range: comparisonTimeRange,
  } = kpiProperties);

  $: schema = validateKPISchema(ctx, kpiProperties);

  $: measure = spec.getMeasureForMetricView(measureName, metricsViewName);
  $: measureValue = useKPITotals(ctx, kpiProperties, $schema.isValid);
  $: measureIsPercentage = $measure?.formatPreset === FormatPreset.PERCENTAGE;

  $: showSparkline = sparkline !== "none";
  $: isSparkRight = sparkline === "right";

  $: showComparison =
    ($showTimeComparison || comparisonTimeRange) && comparisonOptions;
  $: comparisonValue = useKPIComparisonTotal(
    ctx,
    kpiProperties,
    $schema.isValid,
  );
  $: comparisonPercChange =
    $measureValue.data != null && $comparisonValue.data
      ? ($measureValue.data - $comparisonValue.data) / $comparisonValue.data
      : undefined;
  $: globalComparisonLabel =
    $selectedComparisonTimeRange?.name &&
    TIME_COMPARISON[$selectedComparisonTimeRange?.name]?.label;

  $: adjustForPad = topPadding ? 12 : 0;
  $: sparklineHeight = isSparkRight
    ? containerHeight
    : containerHeight -
      (showComparison && $comparisonValue?.data != null ? 100 : 60) -
      adjustForPad;
  $: sparklineWidth = isSparkRight ? containerWidth - 136 : containerWidth - 10;

  $: measureValueFormatter = $measure
    ? createMeasureValueFormatter<null>($measure, "big-number")
    : () => "no data";

  $: measureValueFormatted = $measureValue.data
    ? measureValueFormatter($measureValue.data)
    : "no data";

  $: numberKind = $measure ? numberKindForMeasure($measure) : NumberKind.ANY;

  // $: sparklineData = useKPISparkline(
  //   ctx,
  //   kpiProperties,
  //   $schema.isValid && showSparkline,
  // );
  $: sparkData = $sparklineData?.data?.data || [];

  $: timeAndFilterStore = ctx.canvasEntity.createTimeAndFilterStore(
    metricsViewName,
    {
      componentFilter: kpiProperties.dimension_filters,
      componentComparisonRange: comparisonTimeRange,
    },
  );

  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: ({ instanceId } = $runtime);

  $: sparklineData = createQueryServiceMetricsViewTimeSeries(
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
        // select: (data) => {
        //   return prepareTimeSeries(
        //     data.data || [],
        //     [],
        //     TIME_GRAIN[defaultGrain]?.duration,
        //     timeZone ?? "UTC",
        //   );
        // },
        queryClient: ctx.queryClient,
      },
    },
  );

  $: comparisonData = createQueryServiceMetricsViewTimeSeries(
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
          $schema.isValid &&
          showSparkline &&
          $showTimeComparison,
        // select: (data) => {
        //   return prepareTimeSeries(
        //     data.data || [],
        //     [],
        //     TIME_GRAIN[defaultGrain]?.duration,
        //     timeZone ?? "UTC",
        //   );
        // },
        queryClient: ctx.queryClient,
      },
    },
  );

  $: ({
    timeGrain,
    timeRange: { timeZone, start, end },
    where,
    comparisonRange: { start: comparisonStart, end: comparisonEnd },
  } = $timeAndFilterStore);

  $: interval = Interval.fromDateTimes(
    DateTime.fromISO(start).setZone(timeZone),
    DateTime.fromISO(end).setZone(timeZone),
  );

  $: console.log({ sparkData });

  function getFormattedDiff(comparisonValue: number) {
    if (!$measureValue.data) return "";
    const delta = $measureValue.data - comparisonValue;
    return `${delta >= 0 ? "+" : ""}${measureValueFormatter(delta)}`;
  }

  $: console.log({ timeGrain, timeZone, interval: interval.isValid });
</script>

{#if $schema.isValid}
  {#if measure && !$measureValue.isFetching}
    <div
      class:flex-col={!isSparkRight}
      class="flex gap-2 items-center justify-center overflow-hidden size-full"
    >
      <div class:!items-start={isSparkRight} class="flex flex-col items-center">
        <h2
          class:text-center={!isSparkRight}
          class="font-medium text-sm text-gray-600"
        >
          {$measure?.displayName || measureName}
        </h2>
        <span class="text-3xl font-medium text-gray-600">
          {measureValueFormatted}
        </span>
      </div>

      <div class="size-full">
        {#if $sparklineData.isFetching}
          <div class="flex items-center justify-center w-full h-full">
            <Spinner status={EntityStatus.Running} />
          </div>
        {:else if sparkData.length === 0}
          <div>no data</div>
        {:else if timeGrain && timeZone}
          <Chart
            yMaxPadding={0.1}
            showGrid={false}
            showAxis={false}
            data={[sparkData, $comparisonData?.data?.data ?? [] ?? []]}
            {timeGrain}
            selectedTimeZone={timeZone}
            yAccessor={kpiProperties.measure}
            formatterFunction={measureValueFormatter}
          />
        {/if}
      </div>
    </div>
    <!-- <div
      bind:clientWidth={containerWidth}
      bind:clientHeight={containerHeight}
      class="flex h-full w-full bg-white items-center outline {isSparkRight
        ? 'flex-row'
        : 'flex-col'}"
      class:pt-2={topPadding && !isSparkRight && showSparkline}
      class:justify-center={!showSparkline || !sparkData.length}
    >
      <div
        class="flex flex-col {isSparkRight
          ? 'w-36 justify-center items-start pl-4 line-clamp-2'
          : 'w-full'} {!showSparkline || !isSparkRight ? 'items-center' : ''}"
      >
        <h2 class="measure-label">{$measure?.displayName || measureName}</h2>
        <div class="measure-value">{measureValueFormatted}</div>
        {#if showComparison && $comparisonValue.data}
          <div class="flex items-baseline gap-x-3 text-sm">
            {#if comparisonOptions?.includes("previous") && $comparisonValue.data != null}
              <div role="complementary" class="comparison-value">
                {measureValueFormatter($comparisonValue.data)}
              </div>
            {/if}
            {#if comparisonOptions?.includes("delta")}
              <div role="complementary" class="comparison-value">
                {#if $comparisonValue.data != null}
                  <span
                    class:text-red-500={$measureValue.data &&
                      $measureValue.data - $comparisonValue.data < 0}
                    >{getFormattedDiff($comparisonValue.data)}</span
                  >
                {:else}
                  <span
                    class="ui-copy-disabled-faint italic"
                    style:font-size=".9em">no change</span
                  >
                {/if}
              </div>
            {/if}
            {#if comparisonOptions?.includes("percent_change") && comparisonPercChange != null && !measureIsPercentage}
              <div
                role="complementary"
                class="w-fit font-semibold ui-copy-inactive"
                class:text-red-500={$measureValue.data &&
                  $measureValue.data < 0}
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
      {#if containerHeight && containerWidth && showSparkline && sparkData.length && $selectedTimeRange?.interval}
        <KPISparkline
          {sparkData}
          {measureName}
          {sparklineHeight}
          {sparklineWidth}
          {isSparkRight}
          timeGrain={$selectedTimeRange.interval}
          {measureValueFormatter}
          {numberKind}
        />
      {/if}
    </div> -->
  {:else}
    <div class="flex items-center justify-center w-full h-full">
      <Spinner status={EntityStatus.Running} />
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
    @apply text-sm text-gray-500;
  }

  .comparison-value {
    @apply w-fit max-w-full overflow-hidden;
    @apply font-semibold text-ellipsis text-gray-500;
  }
</style>
