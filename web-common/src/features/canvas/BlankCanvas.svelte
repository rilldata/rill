<script lang="ts">
  import { PlusCircleIcon } from "lucide-svelte";
  import {
    ContextMenu,
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuTrigger,
  } from "@rilldata/web-common/components/context-menu";
  import type { CanvasComponentType } from "./components/types";
  import ChartIcon from "./ChartIcon.svelte";
  import type { ComponentType, SvelteComponent } from "svelte";

  const menuItems: {
    id: CanvasComponentType;
    label: string;
    icon: ComponentType<SvelteComponent>;
  }[] = [
    { id: "markdown", label: "Text", icon: ChartIcon },
    { id: "kpi", label: "KPI", icon: ChartIcon },
    { id: "image", label: "Image", icon: ChartIcon },
    { id: "bar_chart", label: "Chart", icon: ChartIcon },
    { id: "table", label: "Table", icon: ChartIcon },
  ];
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
          on:click={() => console.log(item.id)}
          class="text-gray-700 text-xs"
        >
          <div class="flex flex-row gap-x-2">
            <svelte:component
              this={item.icon}
              size="16px"
              class="text-gray-700"
            />
            <span class="text-gray-700 text-xs font-normal">{item.label}</span>
          </div>
        </ContextMenuItem>
      {/each}
    </ContextMenuContent>
  </ContextMenu>
</div>
