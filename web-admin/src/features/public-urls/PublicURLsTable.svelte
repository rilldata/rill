<script lang="ts">
  import * as Table from "@rilldata/web-common/components/table-shadcn";
  import {
    Render,
    Subscribe,
    createTable,
    createRender,
  } from "svelte-headless-table";
  import { readable } from "svelte/store";
  import { goto } from "$app/navigation";
  import PublicURLsDeleteRow from "./PublicURLsDeleteRow.svelte";

  // TODO: to be added to orval in `web-common/src/runtime-client/gen/index.schemas.ts`
  interface V1MagicAuthToken {
    id?: string;
    secretHash?: string;
    projectId?: string;
    createdOn?: string;
    expiresOn?: string;
    usedOn?: string;
    createdByUserId?: string;
    attributes?: { [key: string]: any };
    metricsView?: string;
    metricsViewFilterJSON?: string;
    metricsViewFields?: string[];
    state?: string;
  }
  export let magicAuthTokens: V1MagicAuthToken[];
  $: console.log("magicAuthTokens: ", magicAuthTokens);

  export let organization: string;
  export let project: string;

  async function handleClickRow(row: any) {
    // TODO: revisit when token secret is available
    // `/${organization}/${project}/magic-link/${token.id}`
    await goto(`/${organization}/${project}/${row.original.metricsView}`);
  }

  const table = createTable(readable(magicAuthTokens));

  const columns = table.createColumns([
    table.column({
      accessor: (token) => token.metricsView,
      header: "Dashboard name",
    }),
    table.column({
      accessor: (token) => token.expiresOn ?? "-",
      header: "Expires on",
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
      cell: ({ value }) => createRender(PublicURLsDeleteRow, { id: value }),
    }),
  ]);

  const { headerRows, pageRows, tableAttrs, tableBodyAttrs } =
    table.createViewModel(columns);
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
          <Table.Row
            {...rowAttrs}
            on:click={async () => await handleClickRow(row)}
          >
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
