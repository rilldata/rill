<script lang="ts">
  import { SimpleDataGraphic } from "@rilldata/web-common/components/data-graphic/elements";
  import { ChunkedLine } from "@rilldata/web-common/components/data-graphic/marks";
  import { extent } from "d3-array";

  import MeasureBigNumber from "@rilldata/web-common/features/dashboards/big-number/MeasureBigNumber.svelte";
  import { useMetaMeasure } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
    MainLineColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import {
    useKPIComparisonTotal,
    useKPISparkline,
    useKPITotals,
  } from "@rilldata/web-common/features/templates/kpi/selector";
  import { KPIProperties } from "@rilldata/web-common/features/templates/types";
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let rendererProperties: KPIProperties;

  const queryClient = useQueryClient();

  $: instanceId = $runtime?.instanceId;
  $: metricViewName = rendererProperties.metric_view;
  $: measureName = rendererProperties.measure;
  $: timeRange = rendererProperties.time_range;
  $: comparisonTimeRange = rendererProperties?.comparison_range;

  $: measure = useMetaMeasure(instanceId, metricViewName, measureName);

  $: measureValue = useKPITotals(
    instanceId,
    metricViewName,
    measureName,
    timeRange,
  );

  $: comparisonValue = useKPIComparisonTotal(
    instanceId,
    metricViewName,
    measureName,
    comparisonTimeRange,
    timeRange,
    queryClient,
  );

  $: sparkline = useKPISparkline(
    instanceId,
    metricViewName,
    measureName,
    timeRange,
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

<div class="flex flex-row">
  {#if $measure.data && $measureValue.data}
    <MeasureBigNumber
      measure={$measure.data}
      value={$measureValue.data}
      withTimeseries={false}
      showComparison
      comparisonValue={$comparisonValue?.data}
      status={$measureValue?.isFetching
        ? EntityStatus.Running
        : EntityStatus.Idle}
      isMeasureExpanded={true}
    />
  {/if}

  <div>
    {#if sparkData.length}
      <SimpleDataGraphic
        height={comparisonTimeRange ? 70 : 65}
        width={200}
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
        vs last {humaniseISODuration(comparisonTimeRange, false)}
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
