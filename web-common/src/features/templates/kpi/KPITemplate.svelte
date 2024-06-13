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
    useKPISparkline,
    useKPITotals,
    useStartEndTime,
  } from "@rilldata/web-common/features/templates/kpi/selector";
  import { KPITemplateT } from "@rilldata/web-common/features/templates/types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let rendererProperties: KPITemplateT;

  const queryClient = useQueryClient();

  $: instanceId = $runtime?.instanceId;
  $: metricViewName = rendererProperties.metric_view;
  $: measureName = rendererProperties.measure;
  $: timeRange = rendererProperties.time_range;

  $: measure = useMetaMeasure(instanceId, metricViewName, measureName);
  $: measureValue = useKPITotals(
    instanceId,
    metricViewName,
    measureName,
    timeRange,
  );
  $: sparkline = useKPISparkline(
    instanceId,
    metricViewName,
    measureName,
    timeRange,
    queryClient,
  );

  $: sparkData = $sparkline?.data || [];
  $: timeRangeExtents = useStartEndTime(instanceId, metricViewName, timeRange);

  $: console.log($timeRangeExtents.data, sparkData);

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
      status={EntityStatus.Idle}
      isMeasureExpanded={true}
    />
  {/if}

  <div class="flex-grow">
    {#if sparkData.length && $timeRangeExtents.data?.start}
      <SimpleDataGraphic
        height={80}
        width={400}
        overflowHidden={false}
        top={10}
        bottom={10}
        right={10}
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
</div>
