<script lang="ts">
  import Table from "$lib/components/table/Table.svelte";
  import TableRow from "$lib/components/table/TableRow.svelte";
  import TableCell from "$lib/components/table/TableCell.svelte";
  import { createEventDispatcher } from "svelte";
  import PreviewTableHeader from "$lib/components/table/PreviewTableHeader.svelte";
  import EditableTableCell from "$lib/components/table/EditableTableCell.svelte";
  import type { ColumnName } from "$lib/components/table/pinnableUtils";
  import { columnIsPinned } from "$lib/components/table/pinnableUtils.js";
  import AddIcon from "$lib/components/icons/AddIcon.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";

  const dispatch = createEventDispatcher();

  export let columnNames: ColumnName[];
  export let selectedColumns: ColumnName[];
  export let rows: any[];
  export let activeIndex: number;
  // export let

  const CellTypeToComponent = {
    editable: EditableTableCell,
  };
</script>

<Table>
  <!-- headers -->
  <TableRow>
    {#each columnNames as { name, type } (name)}
      {@const thisColumnIsPinned = columnIsPinned(name, selectedColumns)}
      <PreviewTableHeader
        {name}
        {type}
        pinned={thisColumnIsPinned}
        on:pin={() => {
          dispatch("pin", { name, type });
        }}
      />
    {/each}
  </TableRow>
  <!-- values -->
  {#each rows as row, index}
    <TableRow hovered={activeIndex === index && activeIndex !== undefined}>
      {#each columnNames as { name, type, cellType, cellComponent } (index + name)}
        {@const comp =
          CellTypeToComponent[cellType] ?? cellComponent ?? TableCell}
        <svelte:component
          this={comp}
          on:inspect={() => {
            dispatch("activeElement", { name, index, value: row[name] });
          }}
          on:change={(evt) => dispatch("change", evt.detail)}
          {name}
          {type}
          {index}
          value={row[name]}
          isNull={row[name] === null}
        />
      {/each}
    </TableRow>
  {/each}
  <TableRow>
    <td
      class="p-2
        pl-4
        pr-4
        border
        border-gray-200"
    >
      <ContextButton on:click={() => dispatch("add")}>
        <AddIcon />
      </ContextButton>
    </td>
  </TableRow>
</Table>
