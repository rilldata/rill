<script lang="ts">
  import Table from "$lib/components/table-editable/Table.svelte";
  import TableRow from "$lib/components/table-editable/TableRow.svelte";
  import TableCell from "$lib/components/table-editable/TableCell.svelte";
  import { createEventDispatcher } from "svelte";
  import PreviewTableHeader from "$lib/components/table-editable/PreviewTableHeader.svelte";
  import { columnIsPinned } from "$lib/components/table-editable/pinnableUtils";
  import AddIcon from "$lib/components/icons/AddIcon.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import type { ColumnConfig } from "$lib/components/table-editable/ColumnConfig";
  import { TableConfig } from "$lib/components/table-editable/TableConfig";

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
    {#each columnNames as columnConfig (columnConfig.name + columnConfig.label)}
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
      {#each columnNames as column (index + column.name + column.label)}
        <TableCell
          on:inspect={() => {
            dispatch("activeElement", {
              name: column.name,
              index,
              value: row[column.name],
            });
          }}
          on:change={(evt) => dispatch("change", evt.detail)}
          on:delete={(evt) => dispatch("delete", evt.detail)}
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
