<script lang="ts">
  import { Meta, Story } from "@storybook/addon-svelte-csf";
  import {
    Render,
    Subscribe,
    createRender,
    createTable,
  } from "svelte-headless-table";
  import { readable } from "svelte/store";
  import * as Table from "..";
  import { Tag } from "../../tag";
  import { todos } from "./data";

  const table = createTable(readable(todos));

  const columns = table.createColumns([
    table.column({
      accessor: (todo) => todo.id,
      header: "ID",
    }),
    table.column({
      accessor: (todo) => todo.title,
      header: "Title",
    }),
    // Example of how to use a custom cell renderer to render a component
    table.column({
      accessor: (todo) => todo.completed,
      header: "Complete",
      cell: ({ value }) => {
        const color = value ? "blue" : "red";
        return createRender(Tag, {
          color,
        }).slot(value ? "Completed" : "Not completed");
      },
    }),
    // Example of how to use a custom cell renderer to format the data
    table.column({
      accessor: (todo) => todo.completedOn,
      header: "Completed on",
      cell: ({ value }) => {
        if (!value) {
          return "-";
        }
        return new Date(value).toLocaleString(undefined, {
          month: "short",
          day: "numeric",
          hour: "numeric",
          minute: "numeric",
        });
      },
    }),
  ]);

  const { headerRows, pageRows, tableAttrs, tableBodyAttrs } =
    table.createViewModel(columns);
</script>

<Meta title="Components/Table" />

<Story name="Basic">
  <div class="rounded-md border">
    <Table.Root {...$tableAttrs}>
      <Table.Header>
        {#each $headerRows as headerRow}
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
  </div></Story
>
