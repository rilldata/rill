<script lang="ts">
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import Chart from "@rilldata/web-common/components/time-series-chart/Chart.svelte";
  import type { ChartDataPoint } from "@rilldata/web-common/components/time-series-chart/types";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import BigNumberTooltipContent from "@rilldata/web-common/features/dashboards/big-number/BigNumberTooltipContent.svelte";
  import { cellInspectorStore } from "@rilldata/web-common/features/dashboards/stores/cell-inspector-store";
  import RangeDisplay from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/RangeDisplay.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { numberPartsToString } from "@rilldata/web-common/lib/number-formatting/utils/number-parts-utils";
  import {
    V1TimeGrain,
    type MetricsViewSpecMeasure,
    type V1MetricsViewAggregationResponse,
    type V1MetricsViewTimeSeriesResponse,
  } from "@rilldata/web-common/runtime-client";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import type { QueryObserverResult } from "@tanstack/svelte-query";
  import { builderActions, getAttrs } from "bits-ui";
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

  let hoveredPoints: ChartDataPoint[] = [];
  let hoveredValue: "primary" | "comparison" | "delta" | "percent" | null =
    null;

  $: measureIsPercentage = measure?.formatPreset === FormatPreset.PERCENTAGE;

  $: measureValueFormatter = measure
    ? createMeasureValueFormatter<null>(measure, "big-number")
    : () => "no data";

  $: measureValueFormatterTooltip = measure
    ? createMeasureValueFormatter<null>(measure, "tooltip")
    : () => "no data";

  $: measureValueFormatterUnabridged = measure
    ? createMeasureValueFormatter<null>(measure, "unabridged")
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

  // Single source of truth for all computed values
  $: computedValues = {
    primary: hoveredPoints?.[0]?.value != null ? currentValue : primaryTotal,
    comparison: comparisonVal,
    delta:
      comparisonVal != null && currentValue != null
        ? currentValue - comparisonVal
        : null,
    percent: comparisonPercChange,
  } as const;

  // Get value based on hover type
  function getValueForType(type: typeof hoveredValue) {
    switch (type) {
      case "comparison":
        return computedValues.comparison;
      case "delta":
        return computedValues.delta;
      case "percent":
        return computedValues.percent;
      default:
        return computedValues.primary;
    }
  }

  function getFormattedDiff(delta: number) {
    return `${delta >= 0 ? "+" : ""}${measureValueFormatter(delta)}`;
  }

  function handleHoverOrFocus(type: typeof hoveredValue) {
    hoveredValue = type;

    const value = getValueForType(type);
    if (value !== undefined && value !== null) {
      cellInspectorStore.updateValue(value.toString());
    }
  }

  function handleLeaveOrBlur() {
    hoveredValue = null;
  }

  $: activeValue = getValueForType(hoveredValue);

  $: tooltipValue = (() => {
    if (hoveredValue === "percent" && computedValues.percent !== null) {
      return numberPartsToString(
        formatMeasurePercentageDifference(computedValues.percent),
      );
    }
    return activeValue !== null && activeValue !== undefined
      ? measureValueFormatterTooltip(activeValue)
      : "no data";
  })();

  $: copyValue =
    computedValues.primary !== null && computedValues.primary !== undefined
      ? measureValueFormatterUnabridged(computedValues.primary)
      : "no data";

  function shiftClickHandler() {
    if (copyValue === "no data") return;
    copyToClipboard(
      copyValue ?? "",
      `copied measure value "${copyValue}" to clipboard`,
    );
  }
</script>

<div class="wrapper" class:spark-right={isSparkRight}>
  <Tooltip.Root>
    <Tooltip.Trigger asChild let:builder>
      <div
        {...getAttrs([builder])}
        use:builderActions={{ builders: [builder] }}
        class="data-wrapper overflow-hidden cursor-pointer"
        style:min-width="{BIG_NUMBER_MIN_WIDTH - adjustment}px"
        aria-label="{measure?.name ?? ''} KPI data"
        role="button"
        tabindex="0"
        on:click={modified({
          shift: shiftClickHandler,
        })}
        on:keydown={(e) => {
          if (e.shiftKey && e.key === "Enter") {
            shiftClickHandler();
          }
        }}
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
          role="button"
          tabindex="0"
          on:mouseover={() => handleHoverOrFocus("primary")}
          on:mouseleave={handleLeaveOrBlur}
          on:focus={() => handleHoverOrFocus("primary")}
          on:blur={handleLeaveOrBlur}
        >
          {#if primaryTotalResult.isError}
            <AlertTriangleIcon class=" text-red-300" size="34px" />
          {:else if primaryTotalResult.isLoading}
            <div class="loading h-6 w-16"></div>
          {:else if primaryTotalResult.data}
            <span class:opacity-50={primaryTotalResult.isFetching}>
              {measureValueFormatter(computedValues.primary)}
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
                <span
                  class="comparison-value"
                  role="button"
                  tabindex="0"
                  on:mouseover={() => handleHoverOrFocus("comparison")}
                  on:mouseleave={handleLeaveOrBlur}
                  on:focus={() => handleHoverOrFocus("comparison")}
                  on:blur={handleLeaveOrBlur}
                >
                  {measureValueFormatter(computedValues.comparison)}
                </span>
              {/if}

              {#if comparisonOptions?.includes("delta")}
                <span
                  class="comparison-value"
                  class:ui-copy-disabled-faint={computedValues.delta === null}
                  class:italic={computedValues.delta === null}
                  class:text-sm={computedValues.delta === null}
                  role="button"
                  tabindex="0"
                  on:mouseover={() => handleHoverOrFocus("delta")}
                  on:mouseleave={handleLeaveOrBlur}
                  on:focus={() => handleHoverOrFocus("delta")}
                  on:blur={handleLeaveOrBlur}
                >
                  {#if computedValues.delta != null}
                    {getFormattedDiff(computedValues.delta)}
                  {:else}
                    no change
                  {/if}
                </span>
              {/if}

              {#if comparisonOptions?.includes("percent_change") && computedValues.percent != null && !measureIsPercentage}
                <span
                  class="w-fit font-semibold text-fg-disabled"
                  class:text-red-500={computedValues.percent < 0}
                  role="button"
                  tabindex="0"
                  on:mouseover={() => handleHoverOrFocus("percent")}
                  on:mouseleave={handleLeaveOrBlur}
                  on:focus={() => handleHoverOrFocus("percent")}
                  on:blur={handleLeaveOrBlur}
                >
                  <PercentageChange
                    color="text-fg-secondary"
                    showPosSign
                    tabularNumber={false}
                    value={formatMeasurePercentageDifference(
                      computedValues.percent,
                    )}
                  />
                </span>
              {/if}
            {/if}
          </div>

          {#if comparisonLabel}
            <p class="text-sm text-fg-secondary break-words">
              vs {comparisonLabel?.toLowerCase()}
            </p>
          {/if}
        {/if}

        {#if !showSparkline && timeGrain && interval.isValid && !hideTimeRange}
          <span class="text-fg-secondary">
            <RangeDisplay {interval} {timeGrain} />
          </span>
        {/if}
      </div>
    </Tooltip.Trigger>

    {#if measure}
      <Tooltip.Content side="top" sideOffset={8}>
        <BigNumberTooltipContent {measure} value={tooltipValue ?? "no data"} />
      </Tooltip.Content>
    {/if}
  </Tooltip.Root>

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
    min-width: 0;
  }

  .measure-name {
    @apply w-full truncate flex-none;
    @apply text-center font-medium text-sm text-fg-primary;
  }

  :global(.dark) .measure-name {
    @apply text-fg-primary;
  }

  .spark-right .measure-name {
    @apply text-left max-w-40;
  }

  .big-number {
    @apply text-3xl font-medium text-fg-primary;
  }

  :global(.dark) .big-number {
    @apply text-fg-primary;
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
    min-width: 0;
  }

  .comparison-value {
    @apply w-fit max-w-full overflow-hidden;
    @apply font-medium text-ellipsis text-fg-secondary;
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
    @apply bg-gray-200 animate-pulse rounded-full;
  }
</style>
