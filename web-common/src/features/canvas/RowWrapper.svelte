<script lang="ts">
  import { MIN_HEIGHT } from "./layout-util";

  export let zIndex: number;
  export let maxWidth: number;
  export let gridTemplate: string;
  export let height: number | undefined = undefined;
  export let heightUnit: string | undefined = undefined;
  export let id: string;
</script>

<section
  {id}
  role="presentation"
  class="w-full grid canvas-row relative h-fit min-h-fit pointer-events-none"
  style:z-index={zIndex}
  style:max-width="{maxWidth}px"
  style:--row-height="{height || MIN_HEIGHT}{heightUnit || 'px'}"
  style:grid-template-columns={gridTemplate}
>
  <slot />
</section>

<style lang="postcss">
  @reference "tailwindcss";

  @reference "tailwindcss";

  .canvas-row {
    grid-auto-rows: auto;
  }

  @container canvas-container (inline-size < 768px) {
    .canvas-row {
      grid-template-columns: repeat(1, 1fr) !important;
    }
  }
</style>
