<script lang="ts">
  import * as Table from "@rilldata/web-common/components/table-shadcn";
  import type {
    V1ModelSplit,
    V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import {
    Render,
    Subscribe,
    createRender,
    createTable,
  } from "svelte-headless-table";
  import { writable } from "svelte/store";
  import TriggerSplit from "./TriggerSplit.svelte";

  export let resource: V1Resource;
  export let splits: V1ModelSplit[];

  const isIncremental = resource.model?.spec?.incremental;

  const splitsStore = writable(splits);
  $: splitsStore.set(splits);

  const table = createTable(splitsStore);

  const columns = table.createColumns([
    table.column({
      header: "Key",
      accessor: (split) => split.key,
      cell: ({ value }) => value?.slice(0, 6) + "...",
    }),
    table.column({
      header: "Data",
      accessor: (split) => split.data,
      cell: ({ value }) => (value ? JSON.stringify(value) : "-"),
    }),
    table.column({
      header: "Executed on",
      accessor: (split) => split.executedOn,
      cell: ({ value }) =>
        value
          ? new Date(value).toLocaleString(undefined, {
              month: "short",
              day: "numeric",
              hour: "numeric",
              minute: "numeric",
              second: "numeric",
              fractionalSecondDigits: 3,
            })
          : "-",
    }),
    table.column({
      accessor: (split) => split.elapsedMs + "ms",
      header: "Elapsed time",
    }),
    table.column({
      header: "Watermark",
      accessor: (split) => split.watermark,
      cell: ({ value }) =>
        value
          ? new Date(value).toLocaleString(undefined, {
              month: "short",
              day: "numeric",
              hour: "numeric",
              minute: "numeric",
              second: "numeric",
              fractionalSecondDigits: 3,
            })
          : "-",
    }),
    table.column({
      header: "Error",
      accessor: (split) => split.error,
      cell: ({ value }) => (value ? JSON.stringify(value) : "-"),
    }),
    ...(isIncremental
      ? [
          table.column({
            header: "",
            accessor: (split) => split,
            cell: ({ value }) =>
              createRender(TriggerSplit, { resource, split: value }),
          }),
        ]
      : []),
  ]);

  const { headerRows, pageRows, tableAttrs, tableBodyAttrs } =
    table.createViewModel(columns);
</script>

<div class="border rounded-md">
  <Table.Root {...$tableAttrs} wrapperClass="max-h-[80vh]">
    <Table.Header
      class="sticky top-0 bg-white z-10 border-b border-gray-200 rounded-t-md"
    >
      {#each $headerRows as headerRow (headerRow.id)}
        <Subscribe rowAttrs={headerRow.attrs()}>
          <Table.Row>
            {#each headerRow.cells as cell (cell.id)}
              <Subscribe attrs={cell.attrs()} let:attrs props={cell.props()}>
                <Table.Head {...attrs}>
                  <Render of={cell.render()} />
                </Table.Head>
              </Subscribe>
            {/each}
          </Table.Row>
        </Subscribe>
      {/each}
    </Table.Header>
    <Table.Body {...$tableBodyAttrs}>
      {#each $pageRows as row (row.id)}
        <Subscribe rowAttrs={row.attrs()} let:rowAttrs>
          <Table.Row {...rowAttrs}>
            {#each row.cells as cell (cell.id)}
              <Subscribe attrs={cell.attrs()} let:attrs>
                <Table.Cell {...attrs}>
                  <Render of={cell.render()} />
                </Table.Cell>
              </Subscribe>
            {/each}
          </Table.Row>
        </Subscribe>
      {/each}
    </Table.Body>
  </Table.Root>
</div>
