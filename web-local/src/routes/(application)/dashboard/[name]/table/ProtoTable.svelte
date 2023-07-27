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

  const managers = getStateManagers();
  const tableData = getTableData(managers);
  let container;

  const options = writable<TableOptions<any>>({
    data: [],
    columns: [],
    getCoreRowModel: getCoreRowModel(),
  });

  // const rowVirtualizer = useVirtualizer({
  //   count: 10000,
  //   getScrollElement: () => parentRef.current,
  //   estimateSize: () => 35,
  //   overscan: 5,
  // })

  $: rowVirtualizer = createVirtualizer({
    count: 10000,
    getScrollElement: () => container,
    estimateSize: () => 35,
    overscan: 5,
  });

  // const { virtualItems: virtualRows, totalSize } = rowVirtualizer
  let virtualRows = [];
  let totalSize = 0;
  let paddingTop = 0;
  let paddingBottom = 0;
  $: {
    virtualRows = $rowVirtualizer.getVirtualItems();
    totalSize = $rowVirtualizer.getTotalSize();

    paddingTop = virtualRows?.length > 0 ? virtualRows?.[0]?.start || 0 : 0;
    paddingBottom =
      virtualRows?.length > 0
        ? totalSize - (virtualRows?.[virtualRows.length - 1]?.end || 0)
        : 0;

    console.log({
      scrollHeight: container?.scrollHeight,
      scrollTop: container?.scrollTop,
      totalSize,
      paddingTop,
      paddingBottom,
      paddingSize: paddingTop + paddingBottom,
      virtualRowCt: virtualRows.length,
      // virtualSizeTotal: virtualRows.length * 35 + (paddingTop + paddingBottom),
    });
  }

  // $: {
  //   console.log({
  //     virtualRows,
  //     totalSize,
  //     paddingTop,
  //     paddingBottom,
  //   });
  // }
  let container2;
  $: rowVirtualizer2 = createVirtualizer({
    count: 10000,
    getScrollElement: () => container2,
    estimateSize: () => 35,
    overscan: 5,
  });

  $: virtualRows2 = $rowVirtualizer2.getVirtualItems();
  $: totalSize2 = $rowVirtualizer2.getTotalSize();

  $: {
    console.log({
      scrollHeight: container2?.scrollHeight,
      scrollTop: container2?.scrollTop,
      virtualRowCt: virtualRows2.length,
    });
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
  <table class="hidden">
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
      style={`height: ${totalSize}px; max-height: ${totalSize}px; overflow: none;`}
    >
      <tbody>
        {#if paddingTop > 0}
          <tr>
            <td style={`height: ${paddingTop}px`} />
          </tr>
        {/if}
        {#each virtualRows as virtualRow}
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
  <div
    bind:this={container2}
    style="height: 200px; overflow: auto;"
    class="border"
  >
    <div style={`width: 100%; position: relative; height: ${totalSize2}px`}>
      {#each virtualRows2 as row}
        <div
          style={`position: absolute; top: 0px; left: 0px; width: 100%; height: ${row.size}px; transform: translateY(${row.start}px)`}
        >
          {row.index}
        </div>
      {/each}
    </div>
  </div>
</div>
