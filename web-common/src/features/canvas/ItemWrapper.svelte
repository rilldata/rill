<script lang="ts">
  import { getInitialHeight, getMinHeight } from "./layout-util";

  export let zIndex: number;
  export let type: string | undefined = undefined;
  // When true, the wrapper sizes to its content rather than the type's fixed
  // initial height. Used for the "add widget" filler so it doesn't reserve a
  // full row's worth of empty space below the button.
  export let fitContent = false;
  // Whether the row has a height of its own, from the YAML or the resize
  // handle. Without one, the component keeps its initial height.
  export let rowHasHeight = false;

  $: expandable =
    type === "kpi_grid" || type === "markdown" || type === "leaderboard";
  $: minHeight = getMinHeight(type, rowHasHeight);
  // A minimum below the initial height fits the component's content alone.
  $: contentOnlyMinHeight = minHeight < getInitialHeight(type);
</script>

<div
  style:z-index={zIndex}
  class:expandable
  class:fit-content={fitContent}
  class:content-only-min-height={contentOnlyMinHeight}
  style:--min-height="{minHeight}px"
>
  <slot />
</div>

<style lang="postcss">
  div {
    @apply p-2.5 relative pointer-events-none size-full;
    container-type: inline-size;
    container-name: component-container;
  }

  .fit-content {
    @apply h-fit;
  }

  .expandable:not(.fit-content) {
    min-height: var(--row-height);
  }

  :not(.expandable):not(.fit-content) {
    height: max(var(--row-height), var(--min-height));
    min-height: 100%;
  }

  /* A title or description needs its own room on top of such a minimum:
     about 40px for the header and the 20px of padding the content gets with it. */
  .content-only-min-height:has(:global(.component-header-container)) {
    height: max(var(--row-height), calc(var(--min-height) + 60px));
  }
</style>
