<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import {
    prettyResourceKind,
    ResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import ResourceErrorMessage from "./ResourceErrorMessage.svelte";
  import { getResourceKindTagColor } from "./display-utils";
  import { flexRender } from "@tanstack/svelte-table";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import RefreshCell from "./RefreshCell.svelte";
  import NameCell from "./NameCell.svelte";
  import ActionsCell from "./ActionsCell.svelte";

  export let data: V1Resource[];
  export let refreshSource: (
    resourceKind: string,
    resourceName: string,
  ) => void;

  function isSource(resource: V1Resource) {
    return resource.meta.name.kind === ResourceKind.Source;
  }

  const columns: ColumnDef<V1Resource, any>[] = [
    {
      accessorKey: "title",
      header: "Type",
      accessorFn: (row) => row.meta.name.kind,
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
      cell: ({ getValue }) =>
        flexRender(NameCell, {
          name: getValue() as string,
        }),
    },
    {
      accessorFn: (row) => row.meta.reconcileError,
      header: "Status",
      cell: ({ row }) =>
        flexRender(ResourceErrorMessage, {
          message: row.original.meta.reconcileError,
          status: row.original.meta.reconcileStatus,
        }),
      meta: {
        marginLeft: "1",
      },
    },
    {
      accessorFn: (row) => row.meta.stateUpdatedOn,
      header: "Last refresh",
      cell: (info) =>
        flexRender(RefreshCell, {
          date: info.getValue() as string,
        }),
    },
    {
      accessorFn: (row) => row.meta.reconcileOn,
      header: "Next refresh",
      cell: (info) =>
        flexRender(RefreshCell, {
          date: info.getValue() as string,
        }),
    },
    {
      accessorKey: "actions",
      header: "",
      cell: ({ row }) =>
        flexRender(ActionsCell, {
          resourceKind: row.original.meta.name.kind,
          resourceName: row.original.meta.name.name,
          isSource: isSource(row.original),
          refreshSource,
        }),
      enableSorting: false,
      meta: {
        widthPercent: 0,
      },
    },
  ];
</script>

<BasicTable
  {data}
  {columns}
  columnLayout="minmax(95px, 108px) minmax(100px, 3fr) 48px minmax(80px, 2fr) minmax(100px, 2fr) "
/>
