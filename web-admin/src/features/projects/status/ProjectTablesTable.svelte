<script lang="ts">
  import VirtualizedTable from "@rilldata/web-common/components/table/VirtualizedTable.svelte";
  import { flexRender, type ColumnDef } from "@tanstack/svelte-table";
  import type {
    V1OlapTableInfo,
    V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import { formatCompactInteger } from "@rilldata/web-common/lib/formatters";
  import ModelSizeCell from "./ModelSizeCell.svelte";
  import NameCell from "./NameCell.svelte";
  import MaterializationCell from "./MaterializationCell.svelte";
  import ModelActionsCell from "./ModelActionsCell.svelte";
  import ResourceErrorMessage from "./ResourceErrorMessage.svelte";

  export let tables: V1OlapTableInfo[] = [];
  export let isView: Map<string, boolean> = new Map();
  export let columnCount: Map<string, number> = new Map();
  export let rowCount: Map<string, number> = new Map();
  export let modelResources: Map<string, V1Resource> = new Map();
  export let onModelInfoClick: (resource: V1Resource) => void = () => {};
  export let onViewPartitionsClick: (resource: V1Resource) => void = () => {};
  export let onRefreshErroredClick: (resource: V1Resource) => void = () => {};
  export let onIncrementalRefreshClick: (
    resource: V1Resource,
  ) => void = () => {};
  export let onFullRefreshClick: (resource: V1Resource) => void = () => {};

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
      accessorFn: (row) => row.name,
      header: "Name",
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
      id: "rows",
      accessorFn: (row) => rowCount.get(row.name ?? ""),
      header: "Rows",
      sortDescFirst: true,
      cell: ({ getValue }) => {
        const value = getValue() as number | undefined;
        return value !== undefined ? formatCompactInteger(value) : "-";
      },
    },
    {
      id: "columns",
      accessorFn: (row) => columnCount.get(row.name ?? ""),
      header: "Columns",
      sortDescFirst: true,
      cell: ({ getValue }) => {
        const value = getValue() as number | undefined;
        return value !== undefined ? String(value) : "-";
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

        let numA = -1;
        if (sizeA && sizeA !== "-1") {
          numA = typeof sizeA === "number" ? sizeA : parseInt(sizeA, 10);
        }

        let numB = -1;
        if (sizeB && sizeB !== "-1") {
          numB = typeof sizeB === "number" ? sizeB : parseInt(sizeB, 10);
        }

        return numB - numA; // Descending
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
        });
      },
    },
  ] as ColumnDef<V1OlapTableInfo, unknown>[];

  $: tableData = tables;
</script>

{#key [isView, columnCount, rowCount, modelResources]}
  <VirtualizedTable
    tableId="project-tables-table"
    data={tableData}
    {columns}
    columnLayout="minmax(95px, 0.5fr) minmax(150px, 2fr) 64px minmax(80px, 0.8fr) minmax(80px, 0.8fr) minmax(100px, 1fr) 56px"
  />
{/key}
