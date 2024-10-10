<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { prettyResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import ResourceErrorMessage from "./ResourceErrorMessage.svelte";
  import {
    getResourceKindTagColor,
    prettyReconcileStatus,
  } from "./display-utils";
  import { flexRender } from "@tanstack/svelte-table";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import { formatDate } from "@rilldata/web-common/components/table/utils";

  export let data: V1Resource[];

  const columns: ColumnDef<V1Resource, any>[] = [
    {
      accessorKey: "title",
      header: "Type",
      enableSorting: false,
      cell: ({ row }) => {
        const prettyKind = prettyResourceKind(row.original.meta.name.kind);
        const color = getResourceKindTagColor(row.original.meta.name.kind);
        return flexRender(Tag, {
          color,
          text: prettyKind,
        });
      },
    },
    {
      accessorFn: (row) => row.meta.name.name,
      header: "Name",
    },
    {
      accessorFn: (row) => row.meta.reconcileStatus,
      header: "Execution status",
      cell: ({ row }) =>
        prettyReconcileStatus(row.original.meta.reconcileStatus),
    },
    {
      accessorFn: (row) => row.meta.reconcileError,
      header: "Error",
      cell: ({ row }) =>
        flexRender(ResourceErrorMessage, {
          message: row.original.meta.reconcileError,
        }),
    },
    {
      accessorFn: (row) => row.meta.stateUpdatedOn,
      header: "Last refresh",
      cell: (info) => {
        if (!info.getValue()) return "-";
        const date = formatDate(info.getValue() as string);
        return date;
      },
    },
    {
      accessorFn: (row) => row.meta.reconcileOn,
      header: "Next refresh",
      cell: (info) => {
        if (!info.getValue()) return "-";
        const date = formatDate(info.getValue() as string);
        return date;
      },
    },
  ];
</script>

<BasicTable {data} {columns} />
