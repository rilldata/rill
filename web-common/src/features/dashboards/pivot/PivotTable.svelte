<script lang="ts">
  import { writable } from "svelte/store";
  import {
    ColumnDef,
    createSvelteTable,
    ExpandedState,
    flexRender,
    getCoreRowModel,
    getExpandedRowModel,
    getFilteredRowModel,
    TableOptions,
  } from "@tanstack/svelte-table";

  import { makeData, type Person } from "./makeData";
  import PivotExpandableCell from "@rilldata/web-common/features/dashboards/pivot/PivotExpandableCell.svelte";

  const defaultData: Person[] = makeData(100, 5, 3);

  const defaultColumns: ColumnDef<Person>[] = [
    {
      header: "Row Headers",
      footer: (props) => props.column.id,
      columns: [
        {
          accessorKey: "firstName",
          header: "First Name",
          cell: ({ row, getValue }) =>
            flexRender(PivotExpandableCell, {
              value: getValue(),
              row,
            }),
          footer: (props) => props.column.id,
        },
      ],
    },
    {
      header: "Info",
      footer: (props) => props.column.id,
      columns: [
        {
          accessorKey: "age",
          header: () => "Age",
          footer: (props) => props.column.id,
        },
        {
          accessorFn: (row) => row.lastName,
          id: "lastName",
          cell: (info) => info.getValue(),
          header: "Last Name",
          footer: (props) => props.column.id,
        },
        {
          header: "More Info",
          columns: [
            {
              accessorKey: "visits",
              header: "Visits",
              footer: (props) => props.column.id,
            },
            {
              accessorKey: "status",
              header: "Status",
              footer: (props) => props.column.id,
            },
            {
              accessorKey: "progress",
              header: "Profile Progress",
              footer: (props) => props.column.id,
            },
          ],
        },
      ],
    },
  ];

  let expanded: ExpandedState = {};

  function handleExpandedChange(updater) {
    expanded = updater(expanded);

    options.update((options) => ({
      ...options,
      state: {
        expanded,
      },
    }));
  }

  const options = writable<TableOptions<Person>>({
    data: defaultData,
    columns: defaultColumns,
    state: {
      expanded,
    },
    onExpandedChange: handleExpandedChange,
    getSubRows: (row) => row.subRows,
    getFilteredRowModel: getFilteredRowModel(),
    getExpandedRowModel: getExpandedRowModel(),
    getCoreRowModel: getCoreRowModel(),
    enableExpanding: true,
  });

  const rerender = () => {
    options.update((options) => ({
      ...options,
      data: makeData(100, 5, 3),
    }));
  };

  const table = createSvelteTable(options);
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
            <td>
              <svelte:component
                this={flexRender(cell.column.columnDef.cell, cell.getContext())}
              />
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
  }

  tr:nth-child(even) {
    background-color: #f9f9f9;
  }

  tr:hover {
    background-color: #e8e8e8;
  }

  thead {
    border-bottom: 2px solid #333;
    /* position: sticky;
    top: 0; */
  }

  td {
    text-align: right; /* or left, depending on content */
  }
</style>
