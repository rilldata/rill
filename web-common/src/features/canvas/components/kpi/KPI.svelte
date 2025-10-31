<script lang="ts">
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import Chart from "@rilldata/web-common/components/time-series-chart/Chart.svelte";
  import RangeDisplay from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/RangeDisplay.svelte";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import {
    V1TimeGrain,
    type MetricsViewSpecMeasure,
    type V1MetricsViewAggregationResponse,
    type V1MetricsViewTimeSeriesResponse,
  } from "@rilldata/web-common/runtime-client";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import type { QueryObserverResult } from "@tanstack/svelte-query";
  import { AlertTriangleIcon } from "lucide-svelte";
  import { Interval } from "luxon";
  import type { KPISpec } from ".";
  import { BIG_NUMBER_MIN_WIDTH } from ".";

  type Query<T> = QueryObserverResult<T, HTTPError>;
  type TimeSeriesQuery = Query<V1MetricsViewTimeSeriesResponse>;
  type AggregationQuery = Query<V1MetricsViewAggregationResponse>;

  export let primaryTotalResult: AggregationQuery;
  export let comparisonTotalResult: AggregationQuery;
  export let primarySparklineResult: TimeSeriesQuery;
  export let comparisonSparklineResult: TimeSeriesQuery;
  export let measure: MetricsViewSpecMeasure | undefined;
  export let timeGrain: V1TimeGrain | undefined;
  export let timeZone: string | undefined;
  export let interval: Interval;
  export let sparkline: KPISpec["sparkline"];
  export let hideTimeRange: boolean | undefined;
  export let comparisonOptions: KPISpec["comparison"];
  export let showTimeComparison: boolean;
  export let hasTimeSeries: boolean | undefined;
  export let comparisonLabel: string | undefined;

  let hoveredPoints: {
    date: Date;
    value: number | null | undefined;
  }[] = [];

  $: measureIsPercentage = measure?.formatPreset === FormatPreset.PERCENTAGE;

  $: measureValueFormatter = measure
    ? createMeasureValueFormatter<null>(measure, "big-number")
    : () => "no data";

  $: showSparkline = hasTimeSeries && sparkline !== "none";
  $: isSparkRight = sparkline === "right";

  $: showComparison = !!comparisonOptions?.length && showTimeComparison;

  $: primaryTotal = (primaryTotalResult?.data?.data?.[0]?.[
    measure?.name ?? ""
  ] ?? null) as number | null;

  $: comparisonTotal = (comparisonTotalResult?.data?.data?.[0]?.[
    measure?.name ?? ""
  ] ?? null) as number | null;

  $: primaryData = primarySparklineResult?.data?.data ?? [];

  $: comparisonData = comparisonSparklineResult?.data?.data ?? [];

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

  $: adjustment = 30 - (comparisonOptions?.length ?? 0) * 10;

  function getFormattedDiff(comparisonValue: number, currentValue: number) {
    const delta = currentValue - comparisonValue;
    return `${delta >= 0 ? "+" : ""}${measureValueFormatter(delta)}`;
  }
</script>

<div class="wrapper" class:spark-right={isSparkRight}>
  <div
    class="data-wrapper overflow-hidden"
    style:min-width="{BIG_NUMBER_MIN_WIDTH - adjustment}px"
    aria-label="{measure?.name ?? ''} KPI data"
  >
    <h2 class="measure-name" title={measure?.displayName || measure?.name}>
      {#if measure?.displayName}
        {measure.displayName}
      {:else if measure?.name}
        {measure.name}
      {:else}
        <div class="loading h-[14px] w-24"></div>
      {/if}
    </h2>

    <div
      class="big-number h-9 grid place-content-center"
      class:hovered-value={hoveredPoints?.[0]?.value != null}
    >
      {#if primaryTotalResult.isError}
        <AlertTriangleIcon class=" text-red-300" size="34px" />
      {:else if primaryTotalResult.isLoading}
        <div class="loading h-6 w-16"></div>
      {:else if primaryTotalResult.data}
        <span class:opacity-50={primaryTotalResult.isFetching}>
          {measureValueFormatter(
            hoveredPoints?.[0]?.value != null ? currentValue : primaryTotal,
          )}
        </span>
      {/if}
    </div>

    {#if showComparison}
      <div class="comparison-value-wrapper">
        {#if comparisonTotalResult.isError}
          <div class="text-red-400">error loading comparison data</div>
        {:else if comparisonTotalResult.isLoading}
          <div class="loading h-[14px] w-6"></div>
          <div class="loading h-[14px] w-6"></div>
          <div class="loading h-[14px] w-6"></div>
        {:else if comparisonTotalResult.data}
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
                value={formatMeasurePercentageDifference(comparisonPercChange)}
              />
            </span>
          {/if}
        {/if}
      </div>

      {#if comparisonLabel}
        <p class="text-sm text-gray-400 break-words">
          vs {comparisonLabel?.toLowerCase()}
        </p>
      {/if}
    {/if}

    {#if !showSparkline && timeGrain && interval.isValid && !hideTimeRange}
      <span class="text-gray-500">
        <RangeDisplay {interval} {timeGrain} />
      </span>
    {/if}
  </div>

  {#if showSparkline}
    <div
      class="sparkline-wrapper"
      class:opacity-50={primarySparklineResult.isFetching}
      class:saturate-0={primarySparklineResult.isFetching}
    >
      {#if primarySparklineResult.isError}
        <AlertTriangleIcon class="text-red-300" size="34px" />
      {:else if primarySparklineResult.isLoading || !timeGrain || !timeZone || !measure?.name}
        <div
          class="size-full mt-2 !bg-theme-50 loading !rounded-md min-h-10"
        ></div>
      {:else if primarySparklineResult.data}
        <Chart
          bind:hoveredPoints
          {primaryData}
          {hideTimeRange}
          secondaryData={showComparison ? comparisonData : []}
          {timeGrain}
          selectedTimeZone={timeZone}
          yAccessor={measure?.name}
          formatterFunction={measureValueFormatter}
        />
      {/if}
    </div>
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply flex flex-col items-center justify-center size-full gap-2;
    container-type: inline-size;
  }

  .wrapper.spark-right {
    @apply flex-row;
  }

  .data-wrapper {
    @apply flex flex-col w-full h-fit justify-center items-center;
    flex: 1 0 auto;
  }

  .spark-right .data-wrapper {
    @apply items-start h-full;
    flex: 0 4 20%;
  }

  .measure-name {
    @apply w-full truncate flex-none;
    @apply text-center font-medium text-sm text-gray-800;
  }

  :global(.dark) .measure-name {
    @apply text-gray-900;
  }

  .spark-right .measure-name {
    @apply text-left max-w-40;
  }

  .big-number {
    @apply text-3xl font-medium text-gray-800;
  }

  :global(.dark) .big-number {
    @apply text-gray-900;
  }

  .hovered-value {
    @apply text-theme-500;
  }

  .comparison-value-wrapper {
    @apply flex items-center gap-x-2 text-sm -mb-[3px] truncate flex-none h-5;
  }

  .sparkline-wrapper {
    @apply size-full flex items-center justify-center flex-shrink min-h-12;
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

  .loading {
    @apply bg-slate-200 animate-pulse rounded-full;
  }
</style>
