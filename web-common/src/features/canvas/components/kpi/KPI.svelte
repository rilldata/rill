<script lang="ts">
  import { SimpleDataGraphic } from "@rilldata/web-common/components/data-graphic/elements";
  import { ChunkedLine } from "@rilldata/web-common/components/data-graphic/marks";
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import { useMetricsViewSpecMeasure } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
    MainLineColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import type { KPIProperties } from "@rilldata/web-common/features/templates/types";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { extent } from "d3-array";
  import {
    useKPIComparisonTotal,
    useKPISparkline,
    useKPITotals,
  } from "./selector";

  export let rendererProperties: V1ComponentSpecRendererProperties;

  const queryClient = useQueryClient();
  let containerWidth: number;
  let containerHeight: number;

  $: instanceId = $runtime?.instanceId;
  $: kpiProperties = rendererProperties as KPIProperties;

  $: ({
    metrics_view: metricsViewName,
    filter: whereSql,
    measure: measureName,
    time_range: timeRange,
    comparison_range: comparisonTimeRange,
  } = kpiProperties);

  $: measureQuery = useMetricsViewSpecMeasure(
    instanceId,
    metricsViewName,
    measureName,
  );

  $: measure = $measureQuery.data;

  $: measureValue = useKPITotals(
    instanceId,
    metricsViewName,
    measureName,
    timeRange.toUpperCase(),
    whereSql,
  );

  $: comparisonValue = useKPIComparisonTotal(
    instanceId,
    metricsViewName,
    measureName,
    comparisonTimeRange?.toUpperCase(),
    timeRange.toUpperCase(),
    whereSql,
    queryClient,
  );

  $: sparkline = useKPISparkline(
    instanceId,
    metricsViewName,
    measureName,
    timeRange.toUpperCase(),
    whereSql,
    queryClient,
  );

  $: sparkData = $sparkline?.data || [];

  const focusedAreaGradient: [string, string] = [
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
  ];

  $: [yMin, yMax] = extent(sparkData, (d) => d[measureName]);
  $: [xMin, xMax] = extent(sparkData, (d) => d["ts"]);

  $: measureValueFormatter = createMeasureValueFormatter<null>(
    measure,
    "big-number",
  );
  $: measureValueFormatted =
    measureValueFormatter($measureValue.data) ?? "no data";

  $: comparisonPercChange =
    $comparisonValue.data &&
    $measureValue.data !== undefined &&
    $measureValue.data !== null
      ? ($measureValue.data - $comparisonValue.data) / $comparisonValue.data
      : undefined;

  $: measureIsPercentage = measure?.formatPreset === FormatPreset.PERCENTAGE;

  function getFormattedDiff(comparisonValue) {
    const delta = $measureValue.data - comparisonValue;
    return `${delta >= 0 ? "+" : ""}${measureValueFormatter(delta)}`;
  }
</script>

<div
  bind:clientWidth={containerWidth}
  bind:clientHeight={containerHeight}
  class="flex flex-col h-full w-full bg-white p-4"
>
  <div class="flex justify-between items-center mb-2">
    <span class="font-semibold text-lg"
      >{measure?.displayName || measureName}</span
    >
    {#if comparisonTimeRange && $comparisonValue.data}
      <span class="text-sm text-gray-500">
        vs last {humaniseISODuration(comparisonTimeRange?.toUpperCase(), false)}
      </span>
    {/if}
  </div>

  <div class="flex justify-between items-center mb-2">
    <span class="text-2xl font-light">{measureValueFormatted}</span>
    {#if $comparisonValue.data}
      <div class="flex items-baseline gap-x-3 text-sm">
        <div
          role="complementary"
          class="w-fit max-w-full overflow-hidden text-ellipsis ui-copy-inactive"
          class:font-semibold={$measureValue.data >= 0}
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
            class="w-fit ui-copy-inactive"
            class:text-red-500={$measureValue.data < 0}
          >
            <PercentageChange
              tabularNumber={false}
              value={formatMeasurePercentageDifference(comparisonPercChange)}
            />
          </div>
        {/if}
      </div>
    {/if}
  </div>

  {#if sparkData.length}
    <SimpleDataGraphic
      height={containerHeight - 90}
      width={containerWidth - 16}
      overflowHidden={false}
      top={10}
      bottom={0}
      right={10}
      left={0}
      {xMin}
      {xMax}
      {yMin}
      {yMax}
    >
      <ChunkedLine
        lineColor={MainLineColor}
        areaGradientColors={focusedAreaGradient}
        data={sparkData}
        xAccessor="ts"
        yAccessor={measureName}
      />
    </SimpleDataGraphic>
  {/if}
</div>

<style lang="postcss">
</style>
