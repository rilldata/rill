<script lang="ts">
  import { PlusCircleIcon } from "lucide-svelte";
  import {
    ContextMenu,
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuTrigger,
  } from "@rilldata/web-common/components/context-menu";
  import type { CanvasComponentType } from "./components/types";
  import { createEventDispatcher } from "svelte";
  import { menuItems } from "./components/menu-items.svelte";

  const dispatch = createEventDispatcher();

  function handleAddComponent(componentType: CanvasComponentType) {
    dispatch("add", { type: componentType });
  }

  function handleButtonClick(event: MouseEvent) {
    const contextMenuEvent = new MouseEvent("contextmenu", {
      bubbles: true,
      clientX: event.clientX,
      clientY: event.clientY,
    });
    event.target?.dispatchEvent(contextMenuEvent);
  }
</script>

<div class="size-full p-4 bg-white">
  <ContextMenu>
    <ContextMenuTrigger>
      <button
        type="button"
        class="blank-canvas-button flex flex-col items-center gap-2 p-8 rounded-[6px] border border-slate-200 w-full"
        on:click={handleButtonClick}
      >
        <PlusCircleIcon class="w-6 h-6 text-slate-500" />
        <span class="text-sm font-medium text-slate-500">Add a component</span>
      </button>
    </ContextMenuTrigger>
    <ContextMenuContent>
      {#each menuItems as item}
        <ContextMenuItem
          on:click={() => handleAddComponent(item.id)}
          class="text-gray-700 text-xs"
        >
          <div class="flex flex-row gap-x-2">
            <svelte:component this={item.icon} />
            <span class="text-gray-700 text-xs font-normal">{item.label}</span>
          </div>
        </ContextMenuItem>
      {/each}
    </ContextMenuContent>
  </ContextMenu>
</div>

<style lang="postcss">
  .blank-canvas-button:hover {
    /* card-hover */
    box-shadow:
      0px 2px 3px 0px rgba(15, 23, 42, 0.03),
      0px 1px 3px 0px rgba(15, 23, 42, 0.04),
      0px 0px 0px 1px rgba(15, 23, 42, 0.06),
      0px 4px 6px 0px rgba(15, 23, 42, 0.09);
  }
</style>
