<script lang="ts">
  import * as Table from "@rilldata/web-admin/components/table-shadcn";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import type {
    V1ReconcileStatus,
    V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import {
    Render,
    Subscribe,
    createRender,
    createTable,
  } from "svelte-headless-table";
  import { readable } from "svelte/store";
  import ResourceErrorMessage from "./ResourceErrorMessage.svelte";

  export let resources: V1Resource[];

  const table = createTable(readable(resources));

  const columns = table.createColumns([
    table.column({
      accessor: (resource) => resource.meta.name.kind,
      header: "Kind",
      cell: ({ value }) => {
        const prettyKind = prettyResourceKind(value);
        const color = getResourceKindTagColor(value);
        return createRender(Tag, {
          color,
        }).slot(prettyKind);
      },
    }),
    table.column({
      accessor: (resource) => resource.meta.name.name,
      header: "Name",
    }),
    table.column({
      accessor: (resource) => resource.meta.reconcileStatus,
      header: "Execution status",
      cell: ({ value }) => prettyReconcileStatus(value),
    }),
    table.column({
      accessor: (resource) => resource.meta.reconcileError,
      id: "error",
      header: "Error",
      cell: ({ value }) =>
        createRender(ResourceErrorMessage, { message: value }),
    }),
    table.column({
      accessor: (resource) => resource.meta.stateUpdatedOn,
      header: "Last refresh",
      cell: ({ value }) =>
        new Date(value).toLocaleString(undefined, {
          month: "short",
          day: "numeric",
          hour: "numeric",
          minute: "numeric",
        }),
    }),
    table.column({
      accessor: (resource) => resource.meta.reconcileOn,
      header: "Next refresh",
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

  function prettyResourceKind(kind: string) {
    return kind.replace(/^rill\.runtime\.v1\./, "");
  }

  function prettyReconcileStatus(status: V1ReconcileStatus) {
    switch (status) {
      case "RECONCILE_STATUS_IDLE":
        return "Idle";
      case "RECONCILE_STATUS_PENDING":
        return "Pending";
      case "RECONCILE_STATUS_RUNNING":
        return "Running";
      case "RECONCILE_STATUS_UNSPECIFIED":
        return "Unspecified";
    }
  }

  function getResourceKindTagColor(kind: string) {
    switch (kind) {
      case "rill.runtime.v1.MetricsView":
        return "blue";
      case "rill.runtime.v1.Model":
        return "green";
      case "rill.runtime.v1.Report":
        return "purple";
      case "rill.runtime.v1.Source":
        return "orange";
      case "rill.runtime.v1.Theme":
        return "yellow";
      default:
        return "gray";
    }
  }
</script>

<section class="flex flex-col gap-y-4">
  <h2 class="text-lg font-medium">Resources</h2>
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
  </div>
</section>
