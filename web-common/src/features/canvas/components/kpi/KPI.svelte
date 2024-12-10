<script lang="ts">
  import { SimpleDataGraphic } from "@rilldata/web-common/components/data-graphic/elements";
  import { ChunkedLine } from "@rilldata/web-common/components/data-graphic/marks";
  import { useMetricsViewSpecMeasure } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
    MainLineColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import type { KPIProperties } from "@rilldata/web-common/features/templates/types";
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

  $: instanceId = $runtime?.instanceId;
  $: kpiProperties = rendererProperties as KPIProperties;

  $: ({
    metrics_view: metricsViewName,
    filter: whereSql,
    measure: measureName,
    time_range: timeRange,
    comparison_range: comparisonTimeRange,
  } = kpiProperties);

  $: measure = useMetricsViewSpecMeasure(
    instanceId,
    metricsViewName,
    measureName,
  );

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
</script>

<div
  bind:clientWidth={containerWidth}
  class="flex flex-row h-full w-full items-center bg-white"
>
  {#if $measure.data && $measureValue.data}
    <span>{$measureValue.data}</span>
  {/if}

  <div>
    {#if sparkData.length}
      <SimpleDataGraphic
        height={comparisonTimeRange ? 70 : 65}
        width={containerWidth - 160}
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

    {#if comparisonTimeRange}
      <div class="comparison-value">
        vs last {humaniseISODuration(comparisonTimeRange.toUpperCase(), false)}
      </div>
    {/if}
  </div>
</div>

<style lang="postcss">
  .comparison-value {
    font-size: 0.8rem;
    @apply ui-copy-muted pl-1 pt-0.5;
  }
</style>
