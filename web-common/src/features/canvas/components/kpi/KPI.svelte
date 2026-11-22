<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import Chart from "@rilldata/web-common/components/time-series-chart/Chart.svelte";
  import type { ChartDataPoint } from "@rilldata/web-common/components/time-series-chart/types";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import BigNumberTooltipContent from "@rilldata/web-common/features/dashboards/big-number/BigNumberTooltipContent.svelte";
  import type { PrecisionChipProps } from "@rilldata/web-common/features/dashboards/exact-measures/exact-measures";
  import PrecisionChip from "@rilldata/web-common/features/dashboards/exact-measures/PrecisionChip.svelte";
  import { measureSupportsTotalsQuery } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
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
  import type { QueryObserverResult } from "@tanstack/svelte-query";
  import { AlertTriangleIcon } from "lucide-svelte";
  import { Interval } from "luxon";
  import type { KPISpec } from ".";
  import { getBigNumberMinWidth, resolveKPISize } from ".";

  type Query<T> = QueryObserverResult<T, Error>;
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
  export let size: KPISpec["size"] = undefined;
  export let hideTimeRange: boolean | undefined;
  export let comparisonOptions: KPISpec["comparison"];
  export let showTimeComparison: boolean;
  export let hasTimeSeries: boolean | undefined;
  export let comparisonLabel: string | undefined;
  // Precision chip rendered after the measure name; undefined when the measure has no exact variant.
  export let precision: PrecisionChipProps | undefined = undefined;

  let hoveredPoints: ChartDataPoint[] = [];
  // The precision chip has its own tooltip; the card tooltip hides while it is hovered.
  let chipHovered = false;
  let hoveredValue: "primary" | "comparison" | "delta" | "percent" | null =
    null;

  $: measureIsPercentage = measure?.formatPreset === FormatPreset.PERCENTAGE;

  // Measures with required dimensions (e.g. a rolling window ordered by the time
  // dimension) have no single total; the provider skips the totals queries and
  // we show an explanatory hint instead.
  $: supportsTotal = !measure || measureSupportsTotalsQuery(measure);

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

  // The size preset; the styles below key off `data-size`.
  $: kpiSize = resolveKPISize(size);

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
  // Capped at the card's width, so the value stays centered in a narrow card.
  $: dataMinWidth = `min(${getBigNumberMinWidth(kpiSize) - adjustment}px, 100%)`;

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
  $: isDeltaPositive =
    computedValues.delta !== null && computedValues.delta > 0;
  $: isDeltaNegative =
    computedValues.delta !== null && computedValues.delta < 0;
  $: lowerIsBetter = measure?.lowerIsBetter ?? false;
  // When comparisonVal < 0, dividing delta by a negative denominator flips the percentage sign,
  // so "positive %" actually means "value went lower". We flip lowerIsBetter to compensate.
  $: lowerIsBetterForPerc =
    comparisonVal != null && comparisonVal < 0 ? !lowerIsBetter : lowerIsBetter;
  $: comparisonDeltaColorClass = (
    lowerIsBetter ? isDeltaNegative : isDeltaPositive
  )
    ? "text-kpi-positive"
    : (lowerIsBetter ? isDeltaPositive : isDeltaNegative)
      ? "text-kpi-negative"
      : "text-fg-secondary";

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
      const formatted =
        type === "percent" && computedValues.percent !== null
          ? numberPartsToString(
              formatMeasurePercentageDifference(computedValues.percent),
            )
          : (measureValueFormatterTooltip(value) ?? null);
      cellInspectorStore.updateValue(value.toString(), formatted);
    }
  }

  function handleLeaveOrBlur() {
    hoveredValue = null;
  }

  $: activeValue = getValueForType(hoveredValue);

  $: tooltipValue = (() => {
    if (!supportsTotal) {
      return m.kpi_no_total();
    }
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

<div class="wrapper" class:spark-right={isSparkRight} data-size={kpiSize}>
  <Tooltip.Root>
    <Tooltip.Trigger>
      {#snippet child({ props })}
        <div
          {...props}
          class="data-wrapper overflow-hidden cursor-pointer"
          style:min-width={dataMinWidth}
          aria-label={m.kpi_data_aria({ measure: measure?.name ?? "" })}
          role="button"
          tabindex="0"
          onclick={modified({
            shift: shiftClickHandler,
          })}
          onkeydown={(e) => {
            if (e.shiftKey && e.key === "Enter") {
              shiftClickHandler();
            }
          }}
        >
          <h2 class="measure-name">
            {#if measure?.displayName || measure?.name}
              {@const label = measure?.displayName || measure?.name}
              <span title={label}>{label}</span>
            {:else}
              <div class="loading h-[14px] w-24"></div>
            {/if}
            {#if precision}
              <span class="inline-flex align-middle relative -top-px ml-1">
                <PrecisionChip
                  {...precision}
                  onHoverChange={(hovered) => (chipHovered = hovered)}
                />
              </span>
            {/if}
          </h2>

          <div
            class="big-number grid place-content-center"
            class:hovered-value={hoveredPoints?.[0]?.value != null}
            role="button"
            tabindex="0"
            onmouseover={() => handleHoverOrFocus("primary")}
            onmouseleave={handleLeaveOrBlur}
            onfocus={() => handleHoverOrFocus("primary")}
            onblur={handleLeaveOrBlur}
          >
            {#if !supportsTotal}
              <span class="text-fg-muted italic text-sm font-normal">
                {m.kpi_no_total()}
              </span>
            {:else if primaryTotalResult.isError}
              <AlertTriangleIcon class=" text-red-300" size="34px" />
            {:else if primaryTotalResult.isLoading || !measure}
              <div class="loading h-6 w-16"></div>
            {:else if primaryTotalResult.data}
              <span class:opacity-50={primaryTotalResult.isFetching}>
                {measureValueFormatter(computedValues.primary)}
              </span>
            {/if}
          </div>

          {#if showComparison && supportsTotal}
            <div class="comparison-value-wrapper">
              {#if comparisonTotalResult.isError}
                <div class="text-red-400">
                  {m.kpi_comparison_load_error()}
                </div>
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
                    onmouseover={() => handleHoverOrFocus("comparison")}
                    onmouseleave={handleLeaveOrBlur}
                    onfocus={() => handleHoverOrFocus("comparison")}
                    onblur={handleLeaveOrBlur}
                  >
                    {measureValueFormatter(computedValues.comparison)}
                  </span>
                {/if}

                {#if comparisonOptions?.includes("delta")}
                  <span
                    class="comparison-value {comparisonDeltaColorClass}"
                    class:ui-copy-disabled-faint={computedValues.delta === null}
                    class:italic={computedValues.delta === null}
                    role="button"
                    tabindex="0"
                    onmouseover={() => handleHoverOrFocus("delta")}
                    onmouseleave={handleLeaveOrBlur}
                    onfocus={() => handleHoverOrFocus("delta")}
                    onblur={handleLeaveOrBlur}
                  >
                    {#if computedValues.delta != null}
                      <span class={comparisonDeltaColorClass}
                        >{getFormattedDiff(computedValues.delta)}</span
                      >
                    {:else}
                      {m.kpi_no_change()}
                    {/if}
                  </span>
                {/if}

                {#if comparisonOptions?.includes("percent_change") && computedValues.percent != null && !measureIsPercentage}
                  <span
                    class="w-fit {(
                      lowerIsBetter ? isDeltaNegative : isDeltaPositive
                    )
                      ? 'font-semibold'
                      : ''} {comparisonDeltaColorClass}"
                    role="button"
                    tabindex="0"
                    onmouseover={() => handleHoverOrFocus("percent")}
                    onmouseleave={handleLeaveOrBlur}
                    onfocus={() => handleHoverOrFocus("percent")}
                    onblur={handleLeaveOrBlur}
                  >
                    <PercentageChange
                      color="text-fg-secondary"
                      showPosSign
                      tabularNumber={false}
                      lowerIsBetter={lowerIsBetterForPerc}
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
                {m.kpi_vs_comparison({
                  comparison: comparisonLabel?.toLowerCase() ?? "",
                })}
              </p>
            {/if}
          {/if}

          {#if !showSparkline && timeGrain && interval.isValid && !hideTimeRange}
            <span class="text-fg-secondary">
              <RangeDisplay {interval} {timeGrain} />
            </span>
          {/if}
        </div>
      {/snippet}
    </Tooltip.Trigger>

    {#if measure && !chipHovered}
      <Tooltip.Content side="top" sideOffset={8}>
        <BigNumberTooltipContent
          {measure}
          value={tooltipValue ?? "no data"}
          note={supportsTotal ? undefined : m.kpi_no_total_note()}
        />
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
    @apply flex flex-col items-center justify-center size-full;
    gap: var(--kpi-gap);
    container-type: inline-size;

    /*
     * Size preset tokens, here for the default "medium". Each preset scales
     * the card's text and spacing together; the grid scales the card padding
     * and minimum width to match (see KPIGrid.svelte and getMinWidth).
     */
    --kpi-value-size: 1.875rem;
    --kpi-value-height: 2.25rem;
    --kpi-text-size: 0.875rem;
    --kpi-text-height: 1.25rem;
    --kpi-label-max-width: 10rem;
    --kpi-gap: 0.5rem;
    --kpi-sparkline-min-height: 3rem;
  }

  .wrapper[data-size="small"] {
    --kpi-value-size: 1.25rem;
    --kpi-value-height: 1.75rem;
    --kpi-text-size: 0.75rem;
    --kpi-text-height: 1rem;
    --kpi-label-max-width: 8rem;
    --kpi-gap: 0.25rem;
    --kpi-sparkline-min-height: 2rem;
  }

  /*
   * The larger values shrink with the card (cqi is relative to `.wrapper`)
   * once it is too narrow for about eight digits, instead of being clipped.
   */
  .wrapper[data-size="large"] {
    --kpi-value-size: min(2.75rem, 20cqi);
    --kpi-value-height: 3.25rem;
    --kpi-text-size: 1rem;
    --kpi-text-height: 1.5rem;
    --kpi-label-max-width: 14rem;
    --kpi-gap: 0.75rem;
    --kpi-sparkline-min-height: 4rem;
    /* Inherited by the time range caption; smaller presets keep the page's size. */
    font-size: 0.875rem;
  }

  .wrapper[data-size="xlarge"] {
    --kpi-value-size: min(4rem, 20cqi);
    --kpi-value-height: 4.5rem;
    --kpi-text-size: 1.125rem;
    --kpi-text-height: 1.75rem;
    --kpi-label-max-width: 18rem;
    --kpi-gap: 1rem;
    --kpi-sparkline-min-height: 5rem;
    font-size: 1rem;
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
    @apply text-center font-medium text-fg-primary;
    font-size: var(--kpi-text-size);
    line-height: var(--kpi-text-height);
  }

  :global(.dark) .measure-name {
    @apply text-fg-primary;
  }

  .spark-right .measure-name {
    @apply text-left;
    max-width: var(--kpi-label-max-width);
  }

  .big-number {
    @apply font-medium text-fg-primary;
    font-size: var(--kpi-value-size);
    line-height: var(--kpi-value-height);
    height: var(--kpi-value-height);
  }

  :global(.dark) .big-number {
    @apply text-fg-primary;
  }

  .hovered-value {
    @apply text-theme-500;
  }

  .comparison-value-wrapper {
    @apply flex items-center gap-x-2 -mb-[3px] truncate flex-none;
    font-size: var(--kpi-text-size);
    line-height: var(--kpi-text-height);
    height: var(--kpi-text-height);
  }

  /* The "vs <comparison>" label under the comparison values. */
  .data-wrapper > p {
    font-size: var(--kpi-text-size);
    line-height: var(--kpi-text-height);
  }

  .sparkline-wrapper {
    @apply size-full flex items-center justify-center flex-shrink;
    min-height: var(--kpi-sparkline-min-height);
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

  /*
   * The larger presets need a wider value column, so their side sparkline
   * gives way earlier. These query the card itself (`.wrapper`).
   */
  @container (inline-size < 320px) {
    [data-size="large"].spark-right .sparkline-wrapper {
      display: none;
    }

    [data-size="large"].spark-right .data-wrapper {
      align-items: center !important;
      flex: 0 2 auto !important;
    }

    [data-size="large"].spark-right .measure-name {
      max-width: 100% !important;
      text-align: center !important;
    }
  }

  @container (inline-size < 400px) {
    [data-size="xlarge"].spark-right .sparkline-wrapper {
      display: none;
    }

    [data-size="xlarge"].spark-right .data-wrapper {
      align-items: center !important;
      flex: 0 2 auto !important;
    }

    [data-size="xlarge"].spark-right .measure-name {
      max-width: 100% !important;
      text-align: center !important;
    }
  }

  .loading {
    @apply bg-gray-200 animate-pulse rounded-full;
  }
</style>
