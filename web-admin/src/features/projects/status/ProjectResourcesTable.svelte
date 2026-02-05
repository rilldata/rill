<script lang="ts">
  import VirtualizedTable from "@rilldata/web-common/components/table/VirtualizedTable.svelte";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    createRuntimeServiceCreateTrigger,
    getRuntimeServiceListResourcesQueryKey,
    V1ReconcileStatus,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import { flexRender } from "@tanstack/svelte-table";
  import ActionsCell from "./ActionsCell.svelte";
  import NameCell from "./NameCell.svelte";
  import RefreshCell from "./RefreshCell.svelte";
  import RefreshResourceConfirmDialog from "./RefreshResourceConfirmDialog.svelte";
  import ResourceErrorMessage from "./ResourceErrorMessage.svelte";

  export let data: V1Resource[];

  let isConfirmDialogOpen = false;
  let dialogResourceName = "";
  let dialogResourceKind = "";
  let dialogRefreshType: "full" | "incremental" = "full";

  let openDropdownResourceKey = "";

  const createTrigger = createRuntimeServiceCreateTrigger();
  const queryClient = useQueryClient();

  const openRefreshDialog = (
    resourceName: string,
    resourceKind: string,
    refreshType: "full" | "incremental",
  ) => {
    dialogResourceName = resourceName;
    dialogResourceKind = resourceKind;
    dialogRefreshType = refreshType;
    isConfirmDialogOpen = true;
  };

  const closeRefreshDialog = () => {
    isConfirmDialogOpen = false;
  };

  const setDropdownOpen = (resourceKey: string, isOpen: boolean) => {
    openDropdownResourceKey = isOpen ? resourceKey : "";
  };

  const isDropdownOpen = (resourceKey: string) => {
    return openDropdownResourceKey === resourceKey;
  };

  const handleRefresh = async () => {
    if (dialogResourceKind === ResourceKind.Model) {
      await $createTrigger.mutateAsync({
        instanceId: $runtime.instanceId,
        data: {
          models: [
            {
              model: dialogResourceName,
              full: dialogRefreshType === "full",
            },
          ],
        },
      });
    } else {
      await $createTrigger.mutateAsync({
        instanceId: $runtime.instanceId,
        data: {
          resources: [{ kind: dialogResourceKind, name: dialogResourceName }],
        },
      });
    }

    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(
        $runtime.instanceId,
        undefined,
      ),
    });

    closeRefreshDialog();
  };

  // Create columns definition as a constant to prevent unnecessary re-creation
  const columns: ColumnDef<V1Resource, any>[] = [
    {
      accessorKey: "title",
      header: "Type",
      accessorFn: (row) => row.meta.name.kind,
      cell: ({ row }) =>
        flexRender(ResourceTypeBadge, {
          kind: row.original.meta.name.kind as ResourceKind,
        }),
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
        // Only hide actions for reconciling rows
        const status = row.original.meta?.reconcileStatus;
        const isRowReconciling =
          status === V1ReconcileStatus.RECONCILE_STATUS_PENDING ||
          status === V1ReconcileStatus.RECONCILE_STATUS_RUNNING;
        if (!isRowReconciling) {
          const resourceKey = `${row.original.meta.name.kind}:${row.original.meta.name.name}`;
          return flexRender(ActionsCell, {
            resourceKind: row.original.meta.name.kind,
            resourceName: row.original.meta.name.name,
            canRefresh:
              row.original.meta.name.kind === ResourceKind.Model ||
              row.original.meta.name.kind === ResourceKind.Source,
            onClickRefreshDialog: openRefreshDialog,
            isDropdownOpen: isDropdownOpen(resourceKey),
            onDropdownOpenChange: (isOpen: boolean) =>
              setDropdownOpen(resourceKey, isOpen),
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

<RefreshResourceConfirmDialog
  bind:open={isConfirmDialogOpen}
  name={dialogResourceName}
  refreshType={dialogRefreshType}
  onRefresh={handleRefresh}
/>
