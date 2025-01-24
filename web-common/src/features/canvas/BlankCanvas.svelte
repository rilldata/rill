<script lang="ts">
  import { PlusCircleIcon } from "lucide-svelte";
  import {
    ContextMenu,
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuTrigger,
  } from "@rilldata/web-common/components/context-menu";
  import type { CanvasComponentType } from "./components/types";
  import type { ComponentType, SvelteComponent } from "svelte";
  import ChartIcon from "./icons/ChartIcon.svelte";
  import TableIcon from "./icons/TableIcon.svelte";
  import TextIcon from "./icons/TextIcon.svelte";
  import BigNumberIcon from "./icons/BigNumberIcon.svelte";
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  const menuItems: {
    id: CanvasComponentType;
    label: string;
    icon: ComponentType<SvelteComponent>;
  }[] = [
    { id: "bar_chart", label: "Chart", icon: ChartIcon },
    { id: "table", label: "Table", icon: TableIcon },
    { id: "markdown", label: "Text", icon: TextIcon },
    { id: "kpi", label: "KPI", icon: BigNumberIcon },
    { id: "image", label: "Image", icon: ChartIcon },
  ];

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
