<script lang="ts">
  import { SimpleDataGraphic } from "@rilldata/web-common/components/data-graphic/elements";
  import { ChunkedLine } from "@rilldata/web-common/components/data-graphic/marks";
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import {
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
    MainLineColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { extent } from "d3-array";
  import type { KPISpec } from ".";
  import {
    useKPIComparisonTotal,
    useKPISparkline,
    useKPITotals,
    validateKPISchema,
  } from "./selector";

  export let rendererProperties: V1ComponentSpecRendererProperties;

  const ctx = getCanvasStateManagers();
  const {
    spec,
    timeControls: { showTimeComparison },
  } = ctx.canvasEntity;

  let containerWidth: number;
  let containerHeight: number;

  $: kpiProperties = rendererProperties as KPISpec;

  $: ({
    metrics_view: metricsViewName,
    measure: measureName,
    sparkline: showSparkline,
    comparison_range: comparisonTimeRange,
  } = kpiProperties);

  $: schema = validateKPISchema(ctx, kpiProperties);
  $: measure = spec.getMeasureForMetricView(measureName, metricsViewName);

  $: measureValue = useKPITotals(ctx, kpiProperties, $schema.isValid);

  $: comparisonValue = useKPIComparisonTotal(
    ctx,
    kpiProperties,
    $schema.isValid,
  );
  $: showComparison = $showTimeComparison || comparisonTimeRange;

  $: sparkline = useKPISparkline(ctx, kpiProperties, $schema.isValid);
  $: sparkData = $sparkline?.data || [];
  $: isEmptySparkline = sparkData.every((y) => y[measureName] === null);

  const focusedAreaGradient: [string, string] = [
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
  ];

  $: [yMin, yMax] = extent(sparkData, (d) => d[measureName]);
  $: [xMin, xMax] = extent(sparkData, (d) => d["ts_position"]);

  $: measureValueFormatter = $measure
    ? createMeasureValueFormatter<null>($measure, "big-number")
    : () => "no data";

  $: measureValueFormatted = $measureValue.data
    ? measureValueFormatter($measureValue.data)
    : "no data";

  $: comparisonPercChange =
    $comparisonValue.data &&
    $measureValue.data !== undefined &&
    $measureValue.data !== null
      ? ($measureValue.data - $comparisonValue.data) / $comparisonValue.data
      : undefined;

  $: measureIsPercentage = $measure?.formatPreset === FormatPreset.PERCENTAGE;

  $: sparklineHeight =
    containerHeight -
    (!showComparison ||
    (comparisonValue && $comparisonValue?.data === undefined)
      ? 80
      : 104);

  function getFormattedDiff(comparisonValue) {
    if (!$measureValue.data) return "";
    const delta = $measureValue.data - comparisonValue;
    return `${delta >= 0 ? "+" : ""}${measureValueFormatter(delta)}`;
  }
</script>

{#if $schema.isValid}
  {#if measure}
    <div
      bind:clientWidth={containerWidth}
      bind:clientHeight={containerHeight}
      class="flex flex-col h-full w-full bg-white pt-4 items-center gap-y-1"
    >
      <div class="measure-label">{$measure?.displayName || measureName}</div>
      <div class="measure-value">{measureValueFormatted}</div>
      {#if showComparison && $comparisonValue.data}
        <div class="flex items-baseline gap-x-3 text-sm">
          <div
            role="complementary"
            class="w-fit max-w-full overflow-hidden text-ellipsis text-gray-500"
            class:font-semibold={$measureValue.data && $measureValue.data >= 0}
          >
            {#if $comparisonValue.data != null}
              {getFormattedDiff($comparisonValue.data)}
            {:else}
              <span class="ui-copy-disabled-faint italic" style:font-size=".9em"
                >no change</span
              >
            {/if}
          </div>
          {#if comparisonPercChange != null && !measureIsPercentage}
            <div
              role="complementary"
              class="w-fit font-semibold ui-copy-inactive"
              class:text-red-500={$measureValue.data && $measureValue.data < 0}
            >
              <PercentageChange
                color="text-gray-500"
                showPosSign
                tabularNumber={false}
                value={formatMeasurePercentageDifference(comparisonPercChange)}
              />
            </div>
          {/if}
          {#if comparisonTimeRange}
            <span class="comparison-range">
              vs last {humaniseISODuration(
                comparisonTimeRange?.toUpperCase(),
                false,
              )}
            </span>
          {/if}
        </div>
      {/if}
      {#if containerHeight && containerWidth && showSparkline && sparkData.length && !isEmptySparkline}
        <SimpleDataGraphic
          height={sparklineHeight}
          width={containerWidth + 10}
          overflowHidden={false}
          top={5}
          bottom={0}
          right={0}
          left={0}
          {xMin}
          {xMax}
          {yMin}
          {yMax}
        >
          <ChunkedLine
            lineOpacity={0.6}
            stopOpacity={0.2}
            lineColor={MainLineColor}
            areaGradientColors={focusedAreaGradient}
            data={sparkData}
            xAccessor="ts"
            yAccessor={measureName}
          />
        </SimpleDataGraphic>
      {/if}
    </div>
  {:else}
    <div class="flex items-center justify-center w-24">
      <Spinner status={EntityStatus.Running} />
    </div>
  {/if}
{:else}
  <div
    class="flex w-full h-full p-2 text-xl bg-white items-center justify-center text-red-500"
  >
    {$schema.error}
  </div>
{/if}

<style lang="postcss">
  .measure-label {
    @apply font-medium text-sm truncate;
    @apply pr-2 text-gray-700;
  }
  .measure-value {
    @apply text-3xl font-medium text-gray-700;
  }
  .comparison-range {
    @apply text-sm text-gray-500;
  }
</style>
