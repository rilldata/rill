<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TableHeader from "./TableHeader.svelte";

  import type { ColumnConfig } from "./ColumnConfig";

  export let columnConfig: ColumnConfig<any>;

  const name = columnConfig.label ?? columnConfig.name;
</script>

<TableHeader>
  <!--
    FIXME: need a principled way of setting table and column width / min-width.
    This style attr delanda est 
  -->
  <div
    style="min-width: 300px;"
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
        <span class="text-ellipsis overflow-hidden whitespace-nowrap">
          {name}
        </span>
      </div>
      <TooltipContent slot="tooltip-content">
        {columnConfig.headerTooltip}
      </TooltipContent>
    </Tooltip>
    <!--
      FIXME: in conversation with Marissa, we decided to remove pins for now,
      but we'll want to revisit our strategy for freezing columns if the table grows.
     -->
    <!-- <Tooltip location="top" alignment="middle" distance={16}>
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
    </Tooltip> -->
  </div>
</TableHeader>
