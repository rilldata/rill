<script lang="ts">
  import DataTypeIcon from "../data-types/DataTypeIcon.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import Pin from "../icons/Pin.svelte";

  export let name: string = "";
  export let type: string = "";
  // export let nullable: boolean = false;
</script>

<div class="header-cell">
  <DataTypeIcon suppressTooltip color={"text-gray-500"} {type} />

  <p class="truncate">{name}</p>

  <Tooltip location="top" alignment="middle" distance={16}>
    <button
      class="transition-colors duration-100 justify-self-end group-hover:block hidden text-gray-900"
      on:click={() => {
        // pinnedColumns.delete(index);
        // pinnedColumns = pinnedColumns;
      }}
    >
      <Pin size="16px" />
    </button>
    <TooltipContent slot="tooltip-content">
      {"unpin this column from the right side of the table"}
    </TooltipContent>
  </Tooltip>
</div>

<style lang="postcss">
  .header-cell {
    @apply flex gap-x-2 px-4 items-center relative size-full;
    @apply bg-gray-100 font-bold;
  }
</style>
