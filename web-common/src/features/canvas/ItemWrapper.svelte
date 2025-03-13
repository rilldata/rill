<script lang="ts">
  import { getInitialHeight } from "./layout-util";

  export let zIndex: number;
  export let type: string | undefined = undefined;

  $: expandable = type === "kpi_grid" || type === "markdown";
  $: minHeight = getInitialHeight(type) + "px";
</script>

<div style:z-index={zIndex} class:expandable style:--min-height={minHeight}>
  <slot />
</div>

<style lang="postcss">
  div {
    @apply p-2.5 relative pointer-events-none size-full;
    container-type: inline-size;
    container-name: component-container;
  }

  .expandable {
    min-height: var(--row-height);
  }

  :not(.expandable) {
    height: max(var(--row-height), var(--min-height));
    min-height: 100%;
  }
</style>
