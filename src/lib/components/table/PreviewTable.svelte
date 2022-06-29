<script lang="ts">
  /**
   * PreviewTable.svelte
   * Use this component to drop into the application.
   * Its goal it so utilize all of the other container components
   * and provide the interactions needed to do things with the table.
   */
  import { slide } from "svelte/transition";
  import { FormattedDataType } from "$lib/components/data-types/";
  import PinnableTable from "./PinnableTable.svelte";
  import { createEventDispatcher } from "svelte";
  import { togglePin } from "$lib/components/table/pinnableUtils";
  import type { ColumnConfig } from "$lib/components/table/ColumnConfig";
  import type { TableConfig } from "$lib/components/table/TableConfig";

  const dispatch = createEventDispatcher();

  export let columnNames: ColumnConfig[];
  export let tableConfig: TableConfig;
  export let rows: any[];

  let visualCellField = undefined;
  let visualCellValue = undefined;
  let visualCellType = undefined;

  let selectedColumns = [];

  let activeIndex;

  function handlePin({ detail: { columnConfig } }) {
    selectedColumns = togglePin(columnConfig, selectedColumns);
  }

  function setActiveElement({ detail: { value, index, name } }) {
    visualCellValue = value;
    visualCellField = name;
    visualCellType = columnNames.find(
      (column) => column.name === visualCellField
    )?.type;
    activeIndex = index;
  }
</script>

<div class="flex relative">
  <PinnableTable
    on:mouseleave={() => {
      visualCellValue = undefined;
      setActiveElement({ detail: {} });
    }}
    on:pin={handlePin}
    on:activeElement={setActiveElement}
    on:change={(evt) => dispatch("change", evt.detail)}
    on:add={() => dispatch("add")}
    {tableConfig}
    {activeIndex}
    {columnNames}
    {selectedColumns}
    {rows}
  />

  {#if selectedColumns.length}
    <div
      class="sticky right-0 z-20 bg-white border border-l-4 border-t-0 border-b-0 border-r-0 border-gray-300"
    >
      <PinnableTable
        on:pin={handlePin}
        on:activeElement={setActiveElement}
        on:change={(evt) => dispatch("change", evt.detail)}
        {tableConfig}
        {activeIndex}
        columnNames={selectedColumns}
        {selectedColumns}
        {rows}
      />
    </div>
  {/if}
</div>

{#if tableConfig.enablePreview && visualCellValue !== undefined}
  <div
    transition:slide={{ duration: 100 }}
    class="sticky bottom-0 left-0 bg-white p-3 border border-t-1 border-gray-200 pointer-events-none z-30 grid grid-flow-col justify-start gap-x-3 items-baseline"
    style:box-shadow="0 -4px 2px 0 rgb(0 0 0 / 0.05)"
  >
    <span class="font-bold pr-5">{visualCellField}</span>
    <FormattedDataType
      value={visualCellValue}
      type={visualCellType}
      isNull={visualCellValue === null}
    />
  </div>
{/if}
