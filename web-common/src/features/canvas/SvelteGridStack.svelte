<!-- Source: https://github.com/SafetZahirovic/SvelteGridStack -->
<!-- Docs: https://github.com/gridstack/gridstack.js/tree/master/doc -->
<script lang="ts">
  import type { GridStack, GridStackNode, GridStackOptions } from "gridstack";
  import { createEventDispatcher, onDestroy, onMount, tick } from "svelte";

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
  const dispatchEvent = createEventDispatcher();

  let gridEl: HTMLDivElement;

  $: options = {
    column: 12,
    resizable: {
      handles: "e,se,s,sw,w",
    },
    animate: false,
    float: true,
    staticGrid: embed,
    // Note: There is no gap property in gridstack.js, so we use margin to set the gap
    // TODO: might need to half the gap for the top and bottom, require special handling
    margin: `${spec?.gapX || defaults.DEFAULT_TOP_BOTTOM_GAP}px ${spec?.gapY || defaults.DEFAULT_LEFT_RIGHT_GAP}px`,
  } as GridStackOptions;

  function handleMouseDown(event: MouseEvent) {
    const target = event.target as HTMLElement;
    const contentEl = target.closest(".grid-stack-item-content");
    const itemEl = contentEl?.querySelector(".grid-stack-item-content-item");

    if (itemEl) {
      const index = itemEl.getAttribute("data-index");
      if (index !== null) {
        dispatchEvent("select", { index: parseInt(index, 10) });
      }
    }
  }

  // Reactive grid to handle changes in items
  $: if (grid) {
    console.log("[SvelteGridStack] Updating grid");

    grid.batchUpdate();

    const currentItems = grid.getGridItems();
    const currentCount = currentItems.length;
    const newCount = items.length;

    console.log("[SvelteGridStack] currentItems", currentItems);
    console.log("[SvelteGridStack] currentCount", currentCount);
    console.log("[SvelteGridStack] newCount", newCount);

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
        const widget = grid.addWidget({
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

  onMount(async () => {
    const { GridStack } = await import("gridstack");

    grid = GridStack.init(options);

    grid.on("added", async (_: Event, nodes: Array<GridStackNode>) => {
      console.log("[SvelteGridStack] added event, nodes:", nodes);

      await tick(); // Wait for Svelte to update the DOM

      nodes.forEach((node) => {
        // Find the correct index by counting existing grid items
        const gridItems = grid.getGridItems();
        const index = gridItems.findIndex((item) => item === node.el);

        console.log("[SvelteGridStack] adding content for index:", index);

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

        console.log("[SvelteGridStack] appended element for widget", {
          index,
          nodeId: node.id,
          element,
          child,
        });
      });
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

    gridEl.addEventListener("mousedown", handleMouseDown);

    grid.load(items);
  });

  onDestroy(() => {
    gridStackEvents.forEach((event) => grid?.off(event));
    gridEl?.removeEventListener("mousedown", handleMouseDown);

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
    /* @apply flex flex-col items-center justify-center; */
    @apply bg-white border border-gray-200 rounded-md shadow-sm;
  }

  :global(.grid-stack-item-content[data-highlight="true"]),
  :global(.ui-draggable-dragging) {
    @apply border-primary-300;
  }
</style>
