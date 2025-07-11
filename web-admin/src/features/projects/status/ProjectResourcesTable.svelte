<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import {
    prettyResourceKind,
    ResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    V1ReconcileStatus,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import ResourceErrorMessage from "./ResourceErrorMessage.svelte";
  import { getResourceKindTagColor } from "./display-utils";
  import { flexRender } from "@tanstack/svelte-table";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import VirtualizedTable from "@rilldata/web-common/components/table/VirtualizedTable.svelte";
  import RefreshCell from "./RefreshCell.svelte";
  import NameCell from "./NameCell.svelte";
  import ActionsCell from "./ActionsCell.svelte";

  export let data: V1Resource[];
  export let isReconciling: boolean;

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
      accessorFn: (row) => row.meta.reconcileStatus,
      header: "Status",
      sortingFn: (rowA, rowB) => {
        // Priority order: Running (highest) -> Pending -> Idle -> Unknown (lowest)
        const getStatusPriority = (status: V1ReconcileStatus) => {
          switch (status) {
            case V1ReconcileStatus.RECONCILE_STATUS_RUNNING:
              return 4;
            case V1ReconcileStatus.RECONCILE_STATUS_PENDING:
              return 3;
            case V1ReconcileStatus.RECONCILE_STATUS_IDLE:
              return 2;
            case V1ReconcileStatus.RECONCILE_STATUS_UNSPECIFIED:
            default:
              return 1;
          }
        };

        return (
          getStatusPriority(rowB.original.meta.reconcileStatus) -
          getStatusPriority(rowA.original.meta.reconcileStatus)
        );
      },
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
      sortDescFirst: true,
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
      cell: ({ row }) => {
        if (!isReconciling) {
          return flexRender(ActionsCell, {
            resourceKind: row.original.meta.name.kind,
            resourceName: row.original.meta.name.name,
            canRefresh:
              row.original.meta.name.kind === ResourceKind.Model ||
              row.original.meta.name.kind === ResourceKind.Source,
          });
        }
      },
      enableSorting: false,
      meta: {
        widthPercent: 0,
      },
    },
  ];

  $: tableData = data.filter(
    (resource) => resource.meta.name.kind !== ResourceKind.Component,
  );
</script>

<VirtualizedTable
  data={tableData}
  {columns}
  columnLayout="minmax(95px, 108px) minmax(100px, 3fr) 48px minmax(80px, 2fr) minmax(100px, 2fr) 56px"
/>
