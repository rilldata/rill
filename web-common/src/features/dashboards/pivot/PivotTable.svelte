<script lang="ts">
  import { writable } from "svelte/store";
  import {
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getExpandedRowModel,
    TableOptions,
  } from "@tanstack/svelte-table";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

  export let data;
  export let columns;

  const stateManagers = getStateManagers();
  const { dashboardStore, metricsViewName, runtime } = stateManagers;

  // const defaultColumns: ColumnDef<Person>[] = [
  //   {
  //     header: "Row Headers",
  //     footer: (props) => props.column.id,
  //     columns: [
  //       {
  //         accessorKey: "firstName",
  //         header: "First Name",
  //         cell: ({ row, getValue }) =>
  //           flexRender(PivotExpandableCell, {
  //             value: getValue(),
  //             row,
  //           }),
  //         footer: (props) => props.column.id,
  //       },
  //     ],
  //   },
  //   {
  //     header: "Info",
  //     footer: (props) => props.column.id,
  //     columns: [
  //       {
  //         accessorKey: "age",
  //         header: () => "Age",
  //         footer: (props) => props.column.id,
  //       },
  //       {
  //         accessorFn: (row) => row.lastName,
  //         id: "lastName",
  //         cell: (info) => info.getValue(),
  //         header: "Last Name",
  //         footer: (props) => props.column.id,
  //       },
  //       {
  //         header: "More Info",
  //         columns: [
  //           {
  //             accessorKey: "visits",
  //             header: "Visits",
  //             footer: (props) => props.column.id,
  //           },
  //           {
  //             accessorKey: "status",
  //             header: "Status",
  //             footer: (props) => props.column.id,
  //           },
  //           {
  //             accessorKey: "progress",
  //             header: "Profile Progress",
  //             footer: (props) => props.column.id,
  //           },
  //         ],
  //       },
  //     ],
  //   },
  // ];

  $: expanded = $dashboardStore?.pivot?.expanded ?? {};

  function handleExpandedChange(updater) {
    expanded = updater(expanded);
    metricsExplorerStore.setPivotExpanded($metricsViewName, expanded);

    options.update((options) => ({
      ...options,
      state: {
        expanded,
      },
    }));
  }

  const options = writable({
    data: data,
    columns: columns,
    state: {
      expanded,
    },
    onExpandedChange: handleExpandedChange,
    getSubRows: (row) => row.subRows,
    getExpandedRowModel: getExpandedRowModel(),
    getCoreRowModel: getCoreRowModel(),
    enableExpanding: true,
  });

  let table = createSvelteTable(options);

  function rerender() {
    // FIXME: Updating data does not update tanstack table
    // console.log("rerender called with data", data);
    options.update((options) => ({
      ...options,
      data: data,
    }));

    table = createSvelteTable(options);
  }

  // Whenever the input data changes, rerender the table
  $: data && rerender();
</script>

<div class="p-2">
  <table>
    <thead>
      {#each $table.getHeaderGroups() as headerGroup}
        <tr>
          {#each headerGroup.headers as header}
            <th colSpan={header.colSpan}>
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
            {@const result =
              typeof cell.column.columnDef.cell === "function"
                ? cell.column.columnDef.cell(cell.getContext())
                : cell.column.columnDef.cell}
            <td>
              {#if result?.component && result?.props}
                <svelte:component this={result.component} {...result.props} />
              {:else if typeof result === "string" || typeof result === "number"}
                {result}
              {:else}
                <!-- flexRender is REALLY slow https://github.com/TanStack/table/issues/4962#issuecomment-1821011742 -->
                <svelte:component
                  this={flexRender(
                    cell.column.columnDef.cell,
                    cell.getContext()
                  )}
                />
              {/if}
            </td>
          {/each}
        </tr>
      {/each}
    </tbody>
  </table>
  <div class="h-4" />
  <button on:click={() => rerender()} class="border p-2"> Rerender </button>
</div>

<style>
  table {
    width: 100%;
    min-width: 300px;
    border-collapse: collapse;
    color: #333;
  }

  tbody {
    border-bottom: 1px solid lightgray;
  }

  th,
  td {
    padding: 10px;
    border: 1px solid #ddd;
    text-align: left;
  }

  th {
    background-color: #f2f2f2;
    font-weight: bold;
    outline: 1px solid #ddd;
  }

  tr:nth-child(even) {
    background-color: #f9f9f9;
  }

  tr:hover {
    background-color: #e8e8e8;
  }

  thead {
    border-bottom: 2px solid #333;
    position: sticky;
    top: 0;
  }

  td {
    text-align: right; /* or left, depending on content */
  }
</style>
