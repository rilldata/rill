<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import TableHeader from "./TableHeader.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import Pin from "$lib/components/icons/Pin.svelte";

  import type { ColumnConfig } from "./ColumnConfig";

  export let pinned = false;
  export let columnConfig: ColumnConfig;

  const dispatch = createEventDispatcher();
  const name = columnConfig.label ?? columnConfig.name;
</script>

<TableHeader>
  <div
    class="
           flex
           items-center
           justify-stretch
           select-none
           gap-x-2
           "
  >
    <Tooltip location="top" alignment="middle" distance={16}>
      <div
        class="w-full pr-5 flex flex-row gap-x-2 items-center cursor-pointer"
      >
        <span class="text-ellipsis overflow-hidden whitespace-nowrap ">
          {name}
        </span>
      </div>
      <TooltipContent slot="tooltip-content">
        {columnConfig.tooltip}
      </TooltipContent>
    </Tooltip>
    <Tooltip location="top" alignment="middle" distance={16}>
      <button
        class:text-gray-900={pinned}
        class:text-gray-400={!pinned}
        class="transition-colors duration-100 justify-self-end"
        on:click={() => {
          dispatch("pin", { columnConfig });
        }}
      >
        <Pin size="16px" />
      </button>
      <TooltipContent slot="tooltip-content">
        {pinned
          ? "unpin this column from the right side of the table"
          : "pin this column to the right side of the table"}
      </TooltipContent>
    </Tooltip>
  </div>
</TableHeader>
