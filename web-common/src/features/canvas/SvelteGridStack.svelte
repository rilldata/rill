<!-- See: https://github.com/SafetZahirovic/SvelteGridStack -->
<script lang="ts">
  import type {
    GridItemHTMLElement,
    GridStack,
    GridStackNode,
    GridStackOptions,
    GridStackWidget,
  } from "gridstack";
  import { createEventDispatcher, onDestroy, onMount } from "svelte";

  import "gridstack/dist/gridstack-extra.min.css";
  import "gridstack/dist/gridstack.min.css";
  import type {
    GridstackCallbackParams,
    GridstackDispatchEvents,
  } from "./types.ts";

  export let items: Array<GridStackWidget>;
  export let opts: GridStackOptions;

  // See: https://github.com/gridstack/gridstack.js/tree/master/doc#events
  const gridStackEvents = [
    "added",
    "change",
    "disable",
    "drag",
    "dragstart",
    "dragstop",
    "dropped",
    "enable",
    "removed",
    "resizestart",
    "resize",
    "resizestop",
  ] as const;

  const dispatch = createEventDispatcher<GridstackDispatchEvents>();

  let gridEl: HTMLDivElement;
  let grid: GridStack;

  onMount(async () => {
    const { GridStack } = await import("GridStack");
    grid = GridStack.init(opts);
    grid.on("added", (_: Event, items: Array<GridStackNode>) => {
      items.forEach((item, index) => {
        const element = gridEl.querySelector(
          `#grid-id-${index}`,
        ) as HTMLDivElement;
        const child = item.el?.firstElementChild;

        if (!child || !element) {
          console.error("Cannot append append element to GridStack");
          return;
        }
        child.appendChild(element);
        element.style.display = "block";
      });
    });

    gridStackEvents.forEach((event) => {
      grid.on(event, (args: GridstackCallbackParams) => {
        dispatch(event, args);
      });
    });

    grid.load(items);
  });

  onDestroy(() => {
    gridStackEvents.forEach((event) => grid?.off(event));
  });
</script>

<div bind:this={gridEl} class="grid-stack">
  {#each items as item, index}
    <div style="display:none" id={`grid-id-${index}`}>
      <slot {index} {item} />
    </div>
  {/each}
</div>

<style lang="postcss">
  .grid-stack {
    @apply bg-white;
  }
  :global(.grid-stack-item-content) {
    @apply flex flex-col items-center justify-center;
    @apply bg-white border border-gray-200 rounded-md shadow-sm;
  }
</style>
