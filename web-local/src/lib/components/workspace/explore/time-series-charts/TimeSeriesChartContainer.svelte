<!-- @component 
A container GraphicContext for the time series in a metrics dashboard.
-->
<script lang="ts">
  import { GraphicContext } from "../../../data-graphic/elements";
  export let start: Date;
  export let end: Date;
</script>

<div class="grid w-max" style:grid-template-columns="140px max-content">
  <GraphicContext
    bottom={4}
    height={125}
    left={24}
    right={45}
    top={4}
    width={380}
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
