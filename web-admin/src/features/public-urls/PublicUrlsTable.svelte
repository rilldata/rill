<script lang="ts">
  import * as Table from "@rilldata/web-common/components/table-shadcn";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { Render, Subscribe, createTable } from "svelte-headless-table";
  import { readable } from "svelte/store";

  export let resources: V1Resource[];

  const table = createTable(readable(resources));

  const columns = table.createColumns([
    table.column({
      accessor: (resource) => resource.meta.name.name,
      header: "Public URL name",
    }),
    table.column({
      accessor: (resource) => resource.meta.name.name,
      header: "Dashboard name",
    }),
    table.column({
      accessor: (resource) => resource.meta.createdOn,
      header: "Created by",
    }),
    table.column({
      accessor: (resource) => resource.meta.createdOn,
      header: "Last used",
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
