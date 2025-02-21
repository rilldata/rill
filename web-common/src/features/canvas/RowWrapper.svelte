<script lang="ts">
  export let zIndex = 50;
  export let maxWidth: number;
  export let gridTemplate: string;
  export let height: number;
  export let rowId: number;
</script>

<section
  id="canvas-row-{rowId}"
  role="presentation"
  class="w-full grid canvas-row relative min-h-fit"
  style:z-index={zIndex}
  style:max-width="{maxWidth}px"
  style:height="{height}px"
  style:grid-template-columns={gridTemplate}
>
  <slot />
</section>

<style lang="postcss">
  .canvas-row {
    grid-auto-rows: auto;
  }
</style>
