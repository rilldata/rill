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
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { createDebouncer } from "@rilldata/web-common/lib/create-debouncer";

  const managers = getStateManagers();
  const tableData = getTableData(managers);
  let container;

  const options = writable<TableOptions<any>>({
    data: [],
    columns: [],
    getCoreRowModel: getCoreRowModel(),
  });

  $: {
    console.log($tableData);
  }

  $: rowVirtualizer = createVirtualizer({
    count: 10000,
    getScrollElement: () => container,
    estimateSize: () => 35,
    overscan: 5,
  });

  let rowVirtualizerD;

  const debounce = createDebouncer();
  $: {
    debounce(() => {
      rowVirtualizerD = $rowVirtualizer;
    }, 0);
  }

  // $: {
  //   console.log(debouncedRowVirtualizer);
  // }

  // const { virtualItems: virtualRows, totalSize } = rowVirtualizer
  let virtualRows = [];
  let totalSize = 0;
  let paddingTop = 0;
  let paddingBottom = 0;
  $: {
    virtualRows = rowVirtualizerD?.getVirtualItems() ?? [];
    totalSize = rowVirtualizerD?.getTotalSize() ?? 0;

    paddingTop = virtualRows?.length > 0 ? virtualRows?.[0]?.start || 0 : 0;
    paddingBottom =
      virtualRows?.length > 0
        ? totalSize - (virtualRows?.[virtualRows.length - 1]?.end || 0)
        : 0;
  }

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

<div class="p-16">
  <table>
    <thead>
      {#each $table.getHeaderGroups() as headerGroup}
        <tr>
          {#each headerGroup.headers as header}
            <th class="border p-2">
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
            <td class="border p-2">
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
  <div
    bind:this={container}
    style="height: 200px; overflow: auto;"
    class="border"
  >
    <table
      class="w-full border-collapse"
      style={`height: ${totalSize}px; max-height: ${totalSize}px; overflow: none; table-layout: fixed`}
    >
      <thead class="sticky top-0 bg-gray-100">
        <tr>
          <th>hello world</th>
        </tr>
      </thead>
      <tbody>
        {#if paddingTop > 0}
          <tr>
            <td style={`height: ${paddingTop}px`} />
          </tr>
        {/if}
        {#each virtualRows as virtualRow (virtualRow.index)}
          <tr>
            <td style="height: 35px" class="border w-full"
              >{virtualRow.index}</td
            >
          </tr>
        {/each}
        {#if paddingBottom > 0}
          <tr>
            <td style={`height: ${paddingBottom}px`} />
          </tr>
        {/if}
      </tbody>
    </table>
  </div>
</div>
