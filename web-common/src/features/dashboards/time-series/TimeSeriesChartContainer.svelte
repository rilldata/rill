<!-- @component 
A container GraphicContext for the time series in a metrics dashboard.
-->
<script lang="ts">
  import { GraphicContext } from "@rilldata/web-common/components/data-graphic/elements";
  import { ScaleType } from "@rilldata/web-common/components/data-graphic/state";
  import { MEASURE_CONFIG } from "../config";
  export let start: Date | undefined;
  export let end: Date | undefined;
  export let workspaceWidth: number;
  export let timeSeriesWidth: number;
  export let enableFullWidth = false;

  const paddingForFullWidth = 80;
  const paddingForSplitView = 30;
</script>

<div class="max-w-full h-fit flex flex-col max-h-full pr-2">
  <GraphicContext
    bottom={4}
    height={enableFullWidth
      ? MEASURE_CONFIG.chart.fullHeight
      : MEASURE_CONFIG.chart.height}
    left={0}
    right={50}
    fontSize={11}
    top={4}
    width={Math.max(
      enableFullWidth
        ? workspaceWidth - paddingForFullWidth
        : timeSeriesWidth - paddingForSplitView,
    ) - MEASURE_CONFIG.bigNumber.widthWithChart}
    xMax={end}
    xMaxTweenProps={{ duration: 400 }}
    xMin={start}
    xMinTweenProps={{ duration: 400 }}
    xType={ScaleType.DATE}
    yType={ScaleType.NUMBER}
    yMin={0}
  >
    <slot />
  </GraphicContext>
</div>
