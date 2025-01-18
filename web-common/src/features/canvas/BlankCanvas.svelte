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
    console.log("[BlankCanvas] handleAddComponent", componentType);
    dispatch("add", { type: componentType });
  }
</script>

<div class="size-full p-2 bg-white">
  <ContextMenu>
    <ContextMenuTrigger>
      <button
        type="button"
        class="flex flex-col items-center gap-2 p-8 rounded-[6px] border border-slate-200 w-full"
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
