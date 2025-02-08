<script lang="ts">
  import { SimpleDataGraphic } from "@rilldata/web-common/components/data-graphic/elements";
  import { WithBisector } from "@rilldata/web-common/components/data-graphic/functional-components";
  import WithRoundToTimegrain from "@rilldata/web-common/components/data-graphic/functional-components/WithRoundToTimegrain.svelte";
  import { ChunkedLine } from "@rilldata/web-common/components/data-graphic/marks";
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import {
    AreaMutedColorGradientLight,
    MainAreaColorGradientDark,
    MainLineColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import MeasureValueMouseover from "@rilldata/web-common/features/dashboards/time-series/MeasureValueMouseover.svelte";
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
  import { TimeRoundingStrategy } from "@rilldata/web-common/lib/time/types";
  import {
    V1TimeGrain,
    type V1ComponentSpecRendererProperties,
  } from "@rilldata/web-common/runtime-client";
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
    timeControls: {
      showTimeComparison,
      selectedComparisonTimeRange,
      selectedTimeRange,
    },
  } = ctx.canvasEntity;

  const focusedAreaGradient: [string, string] = [
    MainAreaColorGradientDark,
    AreaMutedColorGradientLight,
  ];

  let containerWidth: number;
  let containerHeight: number;
  let mouseoverValue: { x: Date; y: number } | undefined = undefined;
  let hovered = false;

  $: kpiProperties = rendererProperties as KPISpec;

  $: ({
    metrics_view: metricsViewName,
    measure: measureName,
    sparkline: showSparkline,
    comparison_range: comparisonTimeRange,
    sparkline_orientation: sparkLineOrientation,
  } = kpiProperties);

  $: schema = validateKPISchema(ctx, kpiProperties);

  $: measure = spec.getMeasureForMetricView(measureName, metricsViewName);
  $: measureValue = useKPITotals(ctx, kpiProperties, $schema.isValid);
  $: measureIsPercentage = $measure?.formatPreset === FormatPreset.PERCENTAGE;

  $: comparisonValue = useKPIComparisonTotal(
    ctx,
    kpiProperties,
    $schema.isValid,
  );
  $: comparisonPercChange =
    $measureValue.data != null && $comparisonValue.data
      ? ($measureValue.data - $comparisonValue.data) / $comparisonValue.data
      : undefined;
  $: showComparison = $showTimeComparison || comparisonTimeRange;
  $: globalComparisonLabel =
    $selectedComparisonTimeRange?.name &&
    TIME_COMPARISON[$selectedComparisonTimeRange?.name]?.label;

  $: isSparkRight = sparkLineOrientation === "right";
  $: sparkline = useKPISparkline(ctx, kpiProperties, $schema.isValid);
  $: sparkData = $sparkline?.data || [];
  $: isEmptySparkline = sparkData.every((y) => y[measureName] === null);

  $: sparklineHeight = isSparkRight
    ? containerHeight
    : containerHeight -
      (showComparison && $comparisonValue?.data != null ? 112 : 72);
  $: sparklineWidth = isSparkRight ? containerWidth - 136 : containerWidth - 10;

  $: [yMin, yMax] = extent(sparkData, (d) => d[measureName]);
  $: [xMin, xMax] = extent(sparkData, (d) => d["ts_position"]);

  $: measureValueFormatter = $measure
    ? createMeasureValueFormatter<null>($measure, "big-number")
    : () => "no data";

  $: measureValueFormatted = $measureValue.data
    ? measureValueFormatter($measureValue.data)
    : "no data";

  $: numberKind = $measure ? numberKindForMeasure($measure) : NumberKind.ANY;
  $: hoveredTime = mouseoverValue?.x instanceof Date && mouseoverValue?.x;

  function getFormattedDiff(comparisonValue: number) {
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
      class="flex h-full w-full bg-white items-center"
      class:flex-col={!isSparkRight}
      class:pt-4={!isSparkRight}
      class:flex-row={isSparkRight}
    >
      <div
        class="flex flex-col {isSparkRight
          ? 'w-36 justify-center items-start pl-4'
          : 'w-full items-center'}"
      >
        <div class="measure-label">{$measure?.displayName || measureName}</div>
        <div class="measure-value">{measureValueFormatted}</div>
        {#if showComparison && $comparisonValue.data}
          <div class="flex items-baseline gap-x-3 text-sm">
            <div
              role="complementary"
              class="w-fit max-w-full overflow-hidden text-ellipsis text-gray-500"
              class:font-semibold={$measureValue.data &&
                $measureValue.data >= 0}
            >
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
            {#if comparisonPercChange != null && !measureIsPercentage}
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
      {#if containerHeight && containerWidth && showSparkline && sparkData.length && !isEmptySparkline}
        <div class={isSparkRight ? "h-full" : "w-full"}>
          <SimpleDataGraphic
            bind:hovered
            bind:mouseoverValue
            height={sparklineHeight}
            width={sparklineWidth}
            overflowHidden={false}
            top={5}
            bottom={0}
            right={0}
            left={16}
            {xMin}
            {xMax}
            {yMin}
            {yMax}
          >
            <ChunkedLine
              lineOpacity={0.75}
              areaEndOffset="75%"
              lineColor={MainLineColor}
              areaGradientColors={focusedAreaGradient}
              data={sparkData}
              xAccessor="ts"
              yAccessor={measureName}
            />
            {#if hoveredTime && mouseoverValue}
              <WithRoundToTimegrain
                strategy={TimeRoundingStrategy.PREVIOUS}
                value={hoveredTime}
                timeGrain={$selectedTimeRange?.interval ||
                  V1TimeGrain.TIME_GRAIN_HOUR}
                let:roundedValue
              >
                <WithBisector
                  data={sparkData}
                  callback={(d) => d["ts"]}
                  value={roundedValue}
                  let:point
                >
                  <MeasureValueMouseover
                    {point}
                    xAccessor="ts"
                    yAccessor={measureName}
                    showComparison={false}
                    mouseoverFormat={measureValueFormatter}
                    {numberKind}
                  />
                </WithBisector>
              </WithRoundToTimegrain>
            {/if}
          </SimpleDataGraphic>
        </div>
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
