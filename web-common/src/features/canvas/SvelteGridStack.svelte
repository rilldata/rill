<!-- Source: https://github.com/SafetZahirovic/SvelteGridStack -->
<!-- Docs: https://github.com/gridstack/gridstack.js/tree/master/doc#events -->
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
  export let options: GridStackOptions;
  export let grid: GridStack;

  // See: https://github.com/gridstack/gridstack.js/tree/master/doc#events
  const gridStackEvents = [
    "added",
    "change",
    "disable",
    "dragstart",
    "drag",
    "dragstop",
    "dropped",
    "enable",
    "removed",
    "resizestart",
    "resize",
    "resizestop",
  ] as const;

  const dispatchGridstackEvent =
    createEventDispatcher<GridstackDispatchEvents>();
  const dispatchEvent = createEventDispatcher();

  let gridEl: HTMLDivElement;

  function handleClick(event: MouseEvent) {
    const target = event.target as HTMLElement;
    const contentEl = target.closest(".grid-stack-item-content");
    const itemEl = contentEl?.querySelector(".grid-stack-item-content-item");
    // console.log("itemEl", itemEl);

    if (itemEl) {
      const index = itemEl.getAttribute("data-index");
      if (index !== null) {
        dispatchEvent("click", { index: parseInt(index, 10) });
      }
    }
  }

  // Only update grid after initial load
  $: if (grid) {
    console.log("Updating grid");

    grid.batchUpdate();

    const currentItems = grid.getGridItems();
    const currentCount = currentItems.length;
    const newCount = items.length;

    // Update existing items
    items.forEach((item, i) => {
      if (i < currentCount) {
        grid.update(currentItems[i], {
          x: item.x,
          y: item.y,
          w: item.w,
          h: item.h,
        });
      } else {
        grid.addWidget({
          ...item,
          content: `<div class="grid-stack-item-content"></div>`,
        });
      }
    });

    // Remove extra widgets if we have fewer items now
    if (currentCount > newCount) {
      currentItems.slice(newCount).forEach((el) => {
        grid.removeWidget(el, true);
      });
    }

    grid.commit();
  }

  onMount(async () => {
    const { GridStack } = await import("gridstack");

    grid = GridStack.init(options);

    grid.on("added", (_: Event, items: Array<GridStackNode>) => {
      items.forEach((item, index) => {
        const element = gridEl.querySelector(
          `#grid-id-${index}`,
        ) as HTMLDivElement;
        const child = item.el?.firstElementChild;

        if (!child || !element) {
          console.error("Cannot append element to GridStack", {
            index,
            element,
            child,
          });
          return;
        }
        child.appendChild(element);
        element.style.display = "block";

        // FOR TESTING
        element.style.border = "1px solid red";
      });
    });

    gridEl.addEventListener("pointerover", (event) => {
      const target = event?.target as HTMLElement;

      const contentEl = target.closest(".grid-stack-item-content");
      if (contentEl) {
        // FIXME: disable data-highlight in preview mode
        contentEl.setAttribute("data-highlight", "true");
      }
    });

    gridEl.addEventListener("pointerout", (event) => {
      const target = event?.target as HTMLElement;

      const contentEl = target.closest(".grid-stack-item-content");
      if (contentEl) {
        contentEl.removeAttribute("data-highlight");
      }
    });

    gridStackEvents.forEach((event) => {
      grid.on(event, (args: GridstackCallbackParams) => {
        dispatchGridstackEvent(event, args);
      });
    });

    gridEl.addEventListener("click", handleClick);

    // Initial load
    grid.load(items);
  });

  onDestroy(() => {
    gridStackEvents.forEach((event) => grid?.off(event));
    gridEl?.removeEventListener("click", handleClick);

    if (grid) {
      grid.removeAll(true);
      grid.destroy(true);
    }
  });
</script>

<div bind:this={gridEl} class="grid-stack">
  {#each items as item, index}
    <div
      style="display:none"
      id={`grid-id-${index}`}
      data-index={index}
      class="grid-stack-item-content-item"
    >
      <slot {index} {item} />
    </div>
  {/each}
</div>

<style lang="postcss">
  .grid-stack {
    @apply bg-white;
  }

  :global(.grid-stack-item-content) {
    @apply rounded-sm bg-white;
  }

  :global(.grid-stack-item-content) {
    @apply flex flex-col items-center justify-center;
    @apply bg-white border border-gray-200 rounded-md shadow-sm;
  }

  :global(.grid-stack-item-content[data-highlight="true"]),
  :global(.ui-draggable-dragging) {
    @apply border-2 border-primary-300 rounded-sm;
  }
</style>
