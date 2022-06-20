<script lang="ts">
  import Table from "$lib/components/table/Table.svelte";
  import TableRow from "$lib/components/table/TableRow.svelte";
  import TableCell from "$lib/components/table/TableCell.svelte";
  import { createEventDispatcher } from "svelte";
  import PreviewTableHeader from "$lib/components/table/PreviewTableHeader.svelte";
  import type { ColumnConfig } from "$lib/components/table/pinnableUtils";
  import { columnIsPinned } from "$lib/components/table/pinnableUtils.js";
  import AddIcon from "$lib/components/icons/AddIcon.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

  const dispatch = createEventDispatcher();

  export let columnNames: ColumnConfig[];
  export let selectedColumns: ColumnConfig[];
  export let rows: any[];
  export let activeIndex: number;
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
      {#each columnNames as column (index + column.name)}
        <TableCell
          on:inspect={() => {
            dispatch("activeElement", {
              name: column.name,
              index,
              value: row[column.name],
            });
          }}
          on:change={(evt) => dispatch("change", evt.detail)}
          {column}
          {index}
          {row}
          validation={column.validation
            ? column.validation(row, row[column.name])
            : ValidationState.OK}
          value={row[column.name]}
          isNull={row[column.name] === null}
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
