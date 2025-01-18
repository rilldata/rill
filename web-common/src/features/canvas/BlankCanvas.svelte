<script lang="ts">
  import { PlusCircleIcon } from "lucide-svelte";
  import {
    ContextMenu,
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuTrigger,
  } from "@rilldata/web-common/components/context-menu";
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
      <ContextMenuItem on:click={() => console.log("Chart")}>
        Add chart
      </ContextMenuItem>
      <ContextMenuItem on:click={() => console.log("Table")}>
        Add table
      </ContextMenuItem>
      <ContextMenuItem on:click={() => console.log("Text")}>
        Add text
      </ContextMenuItem>
    </ContextMenuContent>
  </ContextMenu>
</div>
