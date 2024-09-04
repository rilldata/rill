<script lang="ts">
  import * as Table from "@rilldata/web-common/components/table-shadcn";
  import {
    Render,
    Subscribe,
    createTable,
    createRender,
  } from "svelte-headless-table";
  import { writable } from "svelte/store";
  import { goto } from "$app/navigation";
  import PublicURLsDeleteRow from "./PublicURLsDeleteRow.svelte";
  import type { V1MagicAuthToken } from "@rilldata/web-common/runtime-client";

  export let magicAuthTokens: V1MagicAuthToken[];
  export let organization: string;
  export let project: string;
  export let onDelete: (deletedTokenId: string) => void;

  const magicAuthTokensStore = writable(magicAuthTokens);
  $: {
    magicAuthTokensStore.set(magicAuthTokens);
  }
  const table = createTable(magicAuthTokensStore);

  const columns = table.createColumns([
    table.column({
      accessor: (token) => token.metricsView,
      header: "Dashboard name",
    }),
    table.column({
      accessor: (token) => token.expiresOn,
      header: "Expires on",
      cell: ({ value }) => {
        if (!value) {
          return "-";
        }
        return new Date(value).toLocaleDateString(undefined, {
          year: "numeric",
          month: "short",
          day: "numeric",
          hour: "numeric",
          minute: "numeric",
        });
      },
    }),
    table.column({
      accessor: (token) => token.attributes.name,
      header: "Created by",
    }),
    table.column({
      accessor: (token) => token.usedOn,
      header: "Last used",
      cell: ({ value }) => {
        if (!value) {
          return "-";
        }
        return new Date(value).toLocaleDateString(undefined, {
          year: "numeric",
          month: "short",
          day: "numeric",
          hour: "numeric",
          minute: "numeric",
        });
      },
    }),
    table.column({
      accessor: (token) => token.id,
      header: "",
      cell: ({ value }) =>
        createRender(PublicURLsDeleteRow, {
          id: value,
          onDelete,
        }),
    }),
  ]);

  const { headerRows, pageRows, tableAttrs, tableBodyAttrs } =
    table.createViewModel(columns);

  function handleClickRow(row: any) {
    // TODOL: REVISIT AFTER https://github.com/rilldata/rill-private-issues/issues/642
    // `/${organization}/${project}/magic-link/${token.id}`
    // http://localhost:3000/dev/rill-github-analytics/-/share/rill_mgc_4nLmVj83NhQ4zACJSww5OHhCGCf1CC97sfpfixe6Jfmu4TjkMMvveE
    goto(`/${organization}/${project}/${row.original.metricsView}`);
  }
</script>

<div class="border rounded-md">
  <Table.Root {...$tableAttrs}>
    <Table.Header>
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
          <Table.Row {...rowAttrs} on:click={() => handleClickRow(row)}>
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
