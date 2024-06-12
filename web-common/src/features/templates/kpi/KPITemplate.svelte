<script lang="ts">
  import MeasureBigNumber from "@rilldata/web-common/features/dashboards/big-number/MeasureBigNumber.svelte";
  import { useMetaMeasure } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { useKPITotals } from "@rilldata/web-common/features/templates/kpi/selector";
  import { KPITemplateT } from "@rilldata/web-common/features/templates/types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let rendererProperties: KPITemplateT;

  $: instanceId = $runtime?.instanceId;
  $: metricViewName = rendererProperties.metric_view;
  $: measureName = rendererProperties.measure;
  $: timeRange = rendererProperties.time_range;

  $: measureValue = useKPITotals(
    instanceId,
    metricViewName,
    measureName,
    timeRange,
  );

  $: measure = useMetaMeasure(instanceId, metricViewName, measureName);

  const focusedAreaGradient: [string, string] = [
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
  ];
</script>

{#if $measure.data && $measureValue.data}
  <MeasureBigNumber
    measure={$measure.data}
    value={$measureValue.data}
    withTimeseries={false}
    status={EntityStatus.Idle}
    isMeasureExpanded={true}
  />
{/if}
<!-- 
<SimpleSvgContainer>
  <ChunkedLine
  lineColor={MainLineColor}
  areaGradientColors={focusedAreaGradient}
  delay={0}
  {data}
  {xAccessor}
  {yAccessor}
/>
</SimpleSvgContainer> -->
