<script lang="ts">
  import VirtualizedTable from "@rilldata/web-common/components/table/VirtualizedTable.svelte";
  import { flexRender, type ColumnDef } from "@tanstack/svelte-table";
  import type {
    V1OlapTableInfo,
    V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import { compareSizes } from "./utils";
  import ModelSizeCell from "./ModelSizeCell.svelte";
  import NameCell from "../resource-table/NameCell.svelte";
  import MaterializationCell from "./MaterializationCell.svelte";
  import ModelActionsCell from "./ModelActionsCell.svelte";
  import ResourceErrorMessage from "../resource-table/ResourceErrorMessage.svelte";

  export let tables: V1OlapTableInfo[] = [];
  export let isView: Map<string, boolean> = new Map();
  export let modelResources: Map<string, V1Resource> = new Map();
  export let onModelInfoClick: (resource: V1Resource) => void = () => {};
  export let onViewPartitionsClick: (resource: V1Resource) => void = () => {};
  export let onRefreshErroredClick: (resource: V1Resource) => void = () => {};
  export let onIncrementalRefreshClick: (
    resource: V1Resource,
  ) => void = () => {};
  export let onFullRefreshClick: (resource: V1Resource) => void = () => {};
  export let onViewLogsClick: (name: string) => void = () => {};

  let openDropdownTableName = "";

  $: columns = [
    {
      id: "materialization",
      accessorFn: (row) => isView.get(row.name ?? ""),
      header: "Type",
      cell: ({ row, getValue }) =>
        flexRender(MaterializationCell, {
          isView: getValue() as boolean | undefined,
          physicalSizeBytes: row.original.physicalSizeBytes,
        }),
    },
    {
      id: "modelName",
      accessorFn: (row) => {
        const resource = modelResources.get((row.name ?? "").toLowerCase());
        return resource?.meta?.name?.name ?? row.name ?? "";
      },
      header: "Model Name",
      cell: ({ getValue }) =>
        flexRender(NameCell, {
          name: getValue() as string,
        }),
    },
    {
      id: "tableName",
      accessorFn: (row) => row.name,
      header: "Table Name",
      cell: ({ getValue }) =>
        flexRender(NameCell, {
          name: getValue() as string,
        }),
    },
    {
      id: "status",
      header: "Status",
      accessorFn: (row) => {
        const resource = modelResources.get((row.name ?? "").toLowerCase());
        return resource?.meta?.reconcileStatus;
      },
      cell: ({ row }) => {
        const resource = modelResources.get(
          (row.original.name ?? "").toLowerCase(),
        );
        return flexRender(ResourceErrorMessage, {
          message: resource?.meta?.reconcileError ?? "",
          status:
            resource?.meta?.reconcileStatus ??
            V1ReconcileStatus.RECONCILE_STATUS_UNSPECIFIED,
          testErrors: resource?.model?.state?.testErrors ?? [],
        });
      },
    },
    {
      id: "size",
      accessorFn: (row) => row.physicalSizeBytes,
      header: "Database Size",
      sortDescFirst: true,
      sortingFn: (rowA, rowB) => {
        const sizeA = rowA.getValue("size") as string | number | undefined;
        const sizeB = rowB.getValue("size") as string | number | undefined;
        return compareSizes(sizeA, sizeB);
      },
      cell: ({ getValue }) =>
        flexRender(ModelSizeCell, {
          sizeBytes: getValue() as string | number | undefined,
        }),
    },
    {
      id: "actions",
      header: "",
      enableSorting: false,
      cell: ({ row }) => {
        const tableName = row.original.name ?? "";
        const resource = modelResources.get(tableName.toLowerCase());
        return flexRender(ModelActionsCell, {
          resource,
          isDropdownOpen: openDropdownTableName === tableName,
          onDropdownOpenChange: (isOpen: boolean) => {
            openDropdownTableName = isOpen ? tableName : "";
          },
          onModelInfoClick,
          onViewPartitionsClick,
          onRefreshErroredClick,
          onIncrementalRefreshClick,
          onFullRefreshClick,
          onViewLogsClick,
        });
      },
    },
  ] as ColumnDef<V1OlapTableInfo, unknown>[];
</script>

<VirtualizedTable
  tableId="models-table"
  data={tables}
  {columns}
  columnLayout="minmax(80px, 0.4fr) minmax(120px, 1.5fr) minmax(120px, 1.5fr) 64px minmax(90px, 0.8fr) 56px"
/>
