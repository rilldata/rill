<script lang="ts">
  /**
   * PreviewTable.svelte
   * Use this component to drop into the application.
   * Its goal it so utilize all of the other container components
   * and provide the interactions needed to do things with the table.
   */
  import ColumnGroup from "./ColumnGroup.svelte";
  import { createEventDispatcher } from "svelte";
  import { togglePin } from "./pinnableUtils";
  import type { ColumnConfig } from "./ColumnConfig";

  const dispatch = createEventDispatcher();

  export let columnNames: ColumnConfig<any>[];

  export let rows: any[];
  export let label: string | undefined = undefined;

  let selectedColumns = [];

  function handlePin({ detail: { columnConfig } }) {
    selectedColumns = togglePin(columnConfig, selectedColumns);
  }
</script>

<div class="flex relative">
  <ColumnGroup
    on:pin={handlePin}
    on:change={(evt) => dispatch("change", evt.detail)}
    on:delete={(evt) => dispatch("delete", evt.detail)}
    on:tableResize
    {columnNames}
    {rows}
    {label}
  />

  {#if selectedColumns.length}
    <div
      class="sticky right-0 z-20 bg-white border border-l-4 border-t-0 border-b-0 border-r-0 border-gray-300"
    >
      <ColumnGroup
        on:pin={handlePin}
        on:change={(evt) => dispatch("change", evt.detail)}
        on:delete={(evt) => dispatch("delete", evt.detail)}
        columnNames={selectedColumns}
        {rows}
        {label}
      />
    </div>
  {/if}
</div>
