<script lang="ts">
  /**
   * PreviewTable.svelte
   * Use this component to drop into the application.
   * Its goal it so utilize all of the other container components
   * and provide the interactions needed to do things with the table.
   */
  import PinnableTable from "./PinnableTable.svelte";
  import { createEventDispatcher } from "svelte";
  import { togglePin } from "$lib/components/table-editable/pinnableUtils";
  import type { ColumnConfig } from "$lib/components/table-editable/ColumnConfig";
  import type { TableConfig } from "$lib/components/table-editable/TableConfig";

  const dispatch = createEventDispatcher();

  export let columnNames: ColumnConfig<any>[];
  export let tableConfig: TableConfig;
  export let rows: any[];

  let selectedColumns = [];

  let activeIndex;

  function handlePin({ detail: { columnConfig } }) {
    selectedColumns = togglePin(columnConfig, selectedColumns);
  }
</script>

<div class="flex relative">
  <PinnableTable
    on:pin={handlePin}
    on:change={(evt) => dispatch("change", evt.detail)}
    on:delete={(evt) => dispatch("delete", evt.detail)}
    on:tableResize
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
        on:change={(evt) => dispatch("change", evt.detail)}
        on:delete={(evt) => dispatch("delete", evt.detail)}
        {activeIndex}
        columnNames={selectedColumns}
        {selectedColumns}
        {rows}
      />
    </div>
  {/if}
</div>
