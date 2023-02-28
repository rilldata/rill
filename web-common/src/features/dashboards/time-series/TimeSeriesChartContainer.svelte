<!-- @component 
A container GraphicContext for the time series in a metrics dashboard.
-->
<script lang="ts">
  import { GraphicContext } from "@rilldata/web-common/components/data-graphic/elements";
  import { MEASURE_CONFIG } from "../config";
  export let start: Date;
  export let end: Date;
  export let workspaceWidth: number;
</script>

<div
  class="grid w-max"
  style:grid-template-columns="{MEASURE_CONFIG.bigNumber.widthWithChart}px
  max-content"
>
  <GraphicContext
    bottom={4}
    height={MEASURE_CONFIG.chart.height}
    left={24}
    right={45}
    top={4}
    width={(workspaceWidth >= MEASURE_CONFIG.breakpoint
      ? MEASURE_CONFIG.container.width.full
      : MEASURE_CONFIG.container.width.breakpoint) -
      MEASURE_CONFIG.bigNumber.widthWithChart}
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
