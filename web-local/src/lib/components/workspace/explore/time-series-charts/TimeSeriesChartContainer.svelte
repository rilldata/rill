<!-- @component 
A container GraphicContext for the time series in a metrics dashboard.
-->
<script lang="ts">
  import { GraphicContext } from "@rilldata/web-common/components/data-graphic/elements";
  export let start: Date;
  export let end: Date;
</script>

{#if start && end && !isNaN(start) && !isNaN(end)}
  <div class="grid w-max" style:grid-template-columns="140px max-content">
    <GraphicContext
      width={380}
      height={125}
      left={24}
      right={45}
      top={4}
      bottom={4}
      xMin={start}
      xMax={end}
      yMin={0}
      xType="date"
      yType="number"
      xMinTweenProps={{ duration: 400 }}
      xMaxTweenProps={{ duration: 400 }}
    >
      <slot />
    </GraphicContext>
  </div>
{/if}
