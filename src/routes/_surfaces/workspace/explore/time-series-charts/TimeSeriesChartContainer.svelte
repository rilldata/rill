<!-- @component 
A container GraphicContext for the time series in a metrics dashboard.
-->
<script lang="ts">
  import { GraphicContext } from "$lib/components/data-graphic/elements";
</script>

<div class="grid w-max" style:grid-template-columns="140px max-content">
  <GraphicContext
    width={380}
    height={125}
    left={24}
    right={45}
    top={4}
    bottom={4}
    yMin={0}
    bodyBuffer={8}
    xType="date"
    yType="number"
  >
    <slot />
  </GraphicContext>
</div>
