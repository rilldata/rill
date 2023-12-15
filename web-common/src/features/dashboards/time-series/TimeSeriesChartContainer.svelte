<!-- @component 
A container GraphicContext for the time series in a metrics dashboard.
-->
<script lang="ts">
  import { GraphicContext } from "@rilldata/web-common/components/data-graphic/elements";
  import { MEASURE_CONFIG } from "../config";
  export let start: Date;
  export let end: Date;
  export let workspaceWidth: number;
  export let enableFullWidth = false;

  const paddingForFullWidth = 80;
</script>

<div
  class="grid w-max pr-2"
  style:grid-template-columns="{MEASURE_CONFIG.bigNumber.widthWithChart}px
  max-content"
>
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
        : workspaceWidth >= MEASURE_CONFIG.breakpoint
        ? MEASURE_CONFIG.container.width.full
        : MEASURE_CONFIG.container.width.breakpoint,
      400
    ) - MEASURE_CONFIG.bigNumber.widthWithChart}
    xMax={end}
    xMaxTweenProps={{ duration: 400 }}
    xMin={start}
    xMinTweenProps={{ duration: 400 }}
    xType="date"
    yMin={0}
    yType="number"
  >
    <slot />
  </GraphicContext>
</div>
