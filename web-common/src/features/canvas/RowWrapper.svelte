<script lang="ts">
  export let zIndex = 50;
  export let maxWidth: number;
  export let gridTemplate: string;
  export let height: number = 80;
  export let heightUnit: string = "px";
  export let rowIndex: number;
</script>

<section
  id="canvas-row-{rowIndex}"
  role="presentation"
  class="w-full grid canvas-row relative h-fit min-h-fit pointer-events-none"
  style:z-index={zIndex}
  style:max-width="{maxWidth}px"
  style:--row-height="{height}{heightUnit}"
  style:grid-template-columns={gridTemplate}
>
  <slot />
</section>

<style lang="postcss">
  .canvas-row {
    grid-auto-rows: auto;
  }

  @container canvas-container (inline-size < 768px) {
    .canvas-row {
      grid-template-columns: repeat(1, 1fr) !important;
    }
  }
</style>
