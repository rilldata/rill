<!-- Source: https://github.com/SafetZahirovic/SvelteGridStack -->
<!-- Docs: https://github.com/gridstack/gridstack.js/tree/master/doc -->
<script lang="ts">
  import type { GridStack, GridStackNode, GridStackOptions } from "gridstack";
  import { createEventDispatcher, onDestroy, onMount } from "svelte";

  import "gridstack/dist/gridstack-extra.min.css";
  import "gridstack/dist/gridstack.min.css";
  import type { GridstackDispatchEvents } from "./types.ts";
  import type {
    V1CanvasItem,
    V1CanvasSpec,
  } from "@rilldata/web-common/runtime-client";
  import * as defaults from "./constants";

  export let items: Array<V1CanvasItem>;
  export let grid: GridStack;
  export let embed = false;
  export let spec: V1CanvasSpec;

  $: console.log("[SvelteGridStack] spec", spec);

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
  const dispatch = createEventDispatcher<{
    select: { index: number };
    deselect: void;
  }>();

  let gridEl: HTMLDivElement;

  let columnCount = 12; // Default column count
  let resizeObserver: ResizeObserver;

  // FYI:
  // There could be a race condition where the grid is updated while dragging
  // so we need to avoid updating the grid while dragging
  // Only update the grid when user finishes dragging
  let isDragging = false;

  function updateColumnCount(width: number) {
    if (width < 398) {
      // Small screens/mobile
      columnCount = 1;
    } else {
      // Large screens
      columnCount = 12;
    }
    if (grid) {
      console.log("[SvelteGridStack] Updating columns to:", columnCount);
      grid.column(columnCount);
    }
  }

  function handlePointerDown(event: PointerEvent) {
    const target = event.target as HTMLElement;

    // Selecting whitespace in the canvas
    if (target.classList.contains("grid-stack")) {
      dispatch("deselect");
      return;
    }

    // Handle component selection
    const contentEl = target.closest(".grid-stack-item-content");
    const itemEl = contentEl?.querySelector(".grid-stack-item-content-item");

    if (itemEl) {
      const index = itemEl.getAttribute("data-index");
      if (index !== null) {
        dispatch("select", { index: parseInt(index, 10) });
      }
    }
  }

  // Reactive grid to handle changes in items
  $: if (grid && !isDragging) {
    console.log("[SvelteGridStack] Updating grid");

    grid.batchUpdate();

    const currentItems = grid.getGridItems();
    const currentCount = currentItems.length;
    const newCount = items.length;

    // Update existing items and add new ones
    items.forEach((item, i) => {
      if (i < currentCount) {
        // Update existing widgets
        grid.update(currentItems[i], {
          x: item.x,
          y: item.y,
          w: item.width,
          h: item.height,
        });
      } else {
        // Add new widget
        console.log("[SvelteGridStack] adding new widget at index", i);
        grid.addWidget({
          x: item.x,
          y: item.y,
          w: item.width,
          h: item.height,
          autoPosition: true,
        });
      }
    });

    // Explicitly remove widgets that are no longer in the items array
    if (currentCount > newCount) {
      currentItems.slice(newCount).forEach((el) => {
        grid.removeWidget(el, true);
      });
    }

    grid.commit();
  }

  onMount(() => {
    // Initialize ResizeObserver to watch container width
    resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        updateColumnCount(entry.contentRect.width);
      }
    });

    if (gridEl) {
      resizeObserver.observe(gridEl);
    }
  });

  onMount(async () => {
    const { GridStack } = await import("gridstack");

    grid = GridStack.init(options);

    grid.on("dragstart", () => {
      isDragging = true;
    });

    grid.on("dragstop", () => {
      isDragging = false;
    });

    grid.on("added", async (_: Event, nodes: Array<GridStackNode>) => {
      grid.batchUpdate();

      // Avoid racy, wait for el to be ready for the visible content
      setTimeout(() => {
        nodes.forEach((node) => {
          // Find the correct index by counting existing grid items
          const gridItems = grid.getGridItems();
          const index = gridItems.findIndex((item) => item === node.el);

          const element = gridEl.querySelector(
            `#grid-id-${index}`,
          ) as HTMLDivElement;
          const child = node.el?.firstElementChild;

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
          element.style.width = "100%";
          element.style.height = "100%";
        });

        grid.commit();
      }, 0);
    });

    gridEl.addEventListener("pointerover", (event) => {
      if (!embed) {
        const target = event?.target as HTMLElement;
        const contentEl = target.closest(".grid-stack-item-content");
        if (contentEl) {
          contentEl.setAttribute("data-highlight", "true");
        }
      }
    });

    gridEl.addEventListener("pointerout", (event) => {
      if (!embed) {
        const target = event?.target as HTMLElement;
        const contentEl = target.closest(".grid-stack-item-content");
        if (contentEl) {
          contentEl.removeAttribute("data-highlight");
        }
      }
    });

    gridStackEvents.forEach((event) => {
      grid.on(event, (args: any) => {
        dispatchGridstackEvent(event, args);
      });
    });

    gridEl.addEventListener("pointerdown", handlePointerDown);

    grid.load(items);
  });

  onDestroy(() => {
    gridStackEvents.forEach((event) => grid?.off(event));
    gridEl?.removeEventListener("pointerdown", handlePointerDown);

    if (grid) {
      grid.removeAll(true);
      grid.destroy(true);
    }

    if (resizeObserver) {
      resizeObserver.disconnect();
    }
  });

  $: options = {
    column: columnCount,
    resizable: {
      handles: "e,se,s,sw,w",
    },
    animate: false,
    float: true,
    staticGrid: embed,
    margin: `${spec?.gapX || defaults.DEFAULT_TOP_BOTTOM_GAP}px ${spec?.gapY || defaults.DEFAULT_LEFT_RIGHT_GAP}px`,
  } as GridStackOptions;
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
    @apply border border-gray-200;
    box-shadow:
      0px 2px 3px 0px rgba(15, 23, 42, 0.03),
      0px 1px 3px 0px rgba(15, 23, 42, 0.04),
      0px 0px 0px 1px rgba(15, 23, 42, 0.06);
  }

  :global(.grid-stack-item-content[data-highlight="true"]) {
    box-shadow:
      0px 2px 3px 0px rgba(15, 23, 42, 0.03),
      0px 1px 3px 0px rgba(15, 23, 42, 0.04),
      0px 0px 0px 1px rgba(15, 23, 42, 0.06),
      0px 4px 6px 0px rgba(15, 23, 42, 0.09);
  }

  :global(.canvas-component) {
    border: 2px solid transparent;
    cursor: grab;
  }

  :global(.canvas-component:hover) {
    cursor: pointer !important;
  }

  :global(.canvas-component:active),
  :global(.canvas-component.ui-draggable-dragging) {
    cursor: grabbing !important;
  }

  :global(.canvas-component[data-selected="true"]) {
    border-color: var(--color-primary-300);
  }
</style>
