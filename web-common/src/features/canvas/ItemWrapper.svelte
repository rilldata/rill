<script lang="ts" context="module">
  /**
   * Action for a component's header (title, description and widget filters).
   * Reports the header's rendered height to the wrapper around the component,
   * which keeps that room on top of a minimum that fits the content alone.
   */
  export function reportHeaderHeight(header: HTMLElement) {
    const wrapper = header.closest<HTMLElement>(".canvas-item-wrapper");
    if (!wrapper) return;

    const observer = new ResizeObserver(() => {
      wrapper.style.setProperty(
        "--component-header-height",
        `${header.offsetHeight}px`,
      );
    });
    observer.observe(header);

    return {
      destroy() {
        observer.disconnect();
        wrapper.style.removeProperty("--component-header-height");
      },
    };
  }
</script>

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
  class="canvas-item-wrapper"
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
    height: max(
      var(--row-height),
      calc(var(--min-height) + var(--header-allowance, 0px))
    );
    min-height: 100%;
  }

  /* A minimum that fits the content alone leaves no room for the header,
     which takes its rendered height on top (see reportHeaderHeight). */
  .content-only-min-height {
    --header-allowance: var(--component-header-height, 0px);
  }

  /* Under a title or description the content also gets 20px of padding
     and a 2px border. */
  .content-only-min-height:has(:global(.component-header-container)) {
    --header-allowance: calc(var(--component-header-height, 0px) + 22px);
  }
</style>
