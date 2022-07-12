<script lang="ts">
  import Table from "$lib/components/table-editable/Table.svelte";
  import TableRow from "$lib/components/table-editable/TableRow.svelte";
  import TableRowWithMenu from "$lib/components/table-editable/TableRowWithMenu.svelte";

  import TableCell from "$lib/components/table-editable/TableCell.svelte";

  import { createEventDispatcher } from "svelte";
  import EditableTableHeader from "$lib/components/table-editable/EditableTableHeader.svelte";

  import AddIcon from "$lib/components/icons/Add.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import type { ColumnConfig } from "$lib/components/table-editable/ColumnConfig";
  import type { TableConfig } from "$lib/components/table-editable/TableConfig";

  import { columnIsPinned } from "$lib/components/table-editable/pinnableUtils";

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
      <EditableTableHeader
        {columnConfig}
        pinned={columnIsPinned(columnConfig.name, selectedColumns)}
        on:pin={() => {
          dispatch("pin", { columnConfig });
        }}
      />
    {/each}
  </TableRow>
  <!-- values -->
  {#each rows as row, index}
    <TableRowWithMenu on:delete={() => dispatch("delete", row.id)}>
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
    </TableRowWithMenu>
  {/each}
</Table>
