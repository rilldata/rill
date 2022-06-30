<script lang="ts">
  import Table from "$lib/components/table-editor/Table.svelte";
  import TableRow from "$lib/components/table-editor/TableRow.svelte";
  import TableCell from "$lib/components/table-editor/TableCell.svelte";
  import { createEventDispatcher } from "svelte";
  import PreviewTableHeader from "$lib/components/table-editor/PreviewTableHeader.svelte";
  import { columnIsPinned } from "$lib/components/table-editor/pinnableUtils";
  import AddIcon from "$lib/components/icons/AddIcon.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import type { ColumnConfig } from "$lib/components/table-editor/ColumnConfig";
  import { TableConfig } from "$lib/components/table-editor/TableConfig";

  const dispatch = createEventDispatcher();

  export let tableConfig: TableConfig;
  export let columnNames: ColumnConfig[];
  export let selectedColumns: ColumnConfig[];
  export let rows: any[];
  export let activeIndex: number;
</script>

<Table>
  <!-- headers -->
  <TableRow>
    {#each columnNames as columnConfig (columnConfig.name)}
      {@const thisColumnIsPinned = columnIsPinned(
        columnConfig.name,
        selectedColumns
      )}
      <PreviewTableHeader
        name={columnConfig.label ?? columnConfig.name}
        type={columnConfig.type}
        pinned={thisColumnIsPinned}
        on:pin={() => {
          dispatch("pin", { columnConfig });
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
  {#if tableConfig.enableAdd}
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
  {/if}
</Table>
