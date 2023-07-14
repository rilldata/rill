<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    createSvelteTable,
    ColumnDef,
    TableOptions,
    getCoreRowModel,
    flexRender,
  } from "@tanstack/svelte-table";
  import { getTableData } from "./selectors";
  import { writable } from "svelte/store";

  const managers = getStateManagers();
  const tableData = getTableData(managers);

  const options = writable<TableOptions<any>>({
    data: [],
    columns: [],
    getCoreRowModel: getCoreRowModel(),
  });

  $: {
    const columns = $tableData.meta.map((field) => {
      return {
        // header: either string, function, or component for rendering header
        // cell: either string, function, or component for rendering cell

        // Method for getting underlying value. Can also use accessorFn or accessorKey
        accessorKey: field.name,
      };
    });
    $options.columns = columns;
    $options.data = $tableData.data;
  }

  const table = createSvelteTable(options);
</script>

<div>
  <table>
    <thead>
      {#each $table.getHeaderGroups() as headerGroup}
        <tr>
          {#each headerGroup.headers as header}
            <th>
              {#if !header.isPlaceholder}
                <svelte:component
                  this={flexRender(
                    header.column.columnDef.header,
                    header.getContext()
                  )}
                />
              {/if}
            </th>
          {/each}
        </tr>
      {/each}
    </thead>
    <tbody>
      {#each $table.getRowModel().rows as row}
        <tr>
          {#each row.getVisibleCells() as cell}
            <td>
              <svelte:component
                this={flexRender(cell.column.columnDef.cell, cell.getContext())}
              />
            </td>
          {/each}
        </tr>
      {/each}
    </tbody>
    <tfoot>
      {#each $table.getFooterGroups() as footerGroup}
        <tr>
          {#each footerGroup.headers as header}
            <th>
              {#if !header.isPlaceholder}
                <svelte:component
                  this={flexRender(
                    header.column.columnDef.footer,
                    header.getContext()
                  )}
                />
              {/if}
            </th>
          {/each}
        </tr>
      {/each}
    </tfoot>
  </table>
</div>
