<script lang="ts">
  import VirtualizedTable from "@rilldata/web-common/components/table/VirtualizedTable.svelte";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    createRuntimeServiceCreateTriggerMutation,
    getRuntimeServiceListResourcesQueryKey,
    V1ReconcileStatus,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useQueryClient } from "@tanstack/svelte-query";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import { flexRender } from "@tanstack/svelte-table";
  import ActionsCell from "./ActionsCell.svelte";
  import NameCell from "./NameCell.svelte";
  import RefreshCell from "./RefreshCell.svelte";
  import RefreshErroredPartitionsDialog from "../tables/RefreshErroredPartitionsDialog.svelte";
  import RefreshResourceConfirmDialog from "./RefreshResourceConfirmDialog.svelte";
  import ResourceErrorMessage from "./ResourceErrorMessage.svelte";
  import ResourceSpecDialog from "./ResourceSpecDialog.svelte";

  export let data: V1Resource[];

  let isConfirmDialogOpen = false;
  let dialogResourceName = "";
  let dialogResourceKind = "";
  let dialogRefreshType: "full" | "incremental" = "full";

  let isSpecDialogOpen = false;
  let specResourceName = "";
  let specResourceKind = "";
  let specResource: V1Resource | undefined = undefined;

  let isErroredPartitionsDialogOpen = false;
  let erroredPartitionsModelName = "";

  let openDropdownResourceKey = "";

  const runtimeClient = useRuntimeClient();
  const createTrigger =
    createRuntimeServiceCreateTriggerMutation(runtimeClient);
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

  const openSpecDialog = (
    resourceName: string,
    resourceKind: string,
    resource: V1Resource,
  ) => {
    specResourceName = resourceName;
    specResourceKind = resourceKind;
    specResource = resource;
    isSpecDialogOpen = true;
  };

  const setDropdownOpen = (resourceKey: string, isOpen: boolean) => {
    openDropdownResourceKey = isOpen ? resourceKey : "";
  };

  const isDropdownOpen = (resourceKey: string) => {
    return openDropdownResourceKey === resourceKey;
  };

  const openRefreshErroredPartitionsDialog = (resourceName: string) => {
    erroredPartitionsModelName = resourceName;
    isErroredPartitionsDialogOpen = true;
  };

  const handleRefreshErroredPartitions = async () => {
    await $createTrigger.mutateAsync({
      models: [
        { model: erroredPartitionsModelName, allErroredPartitions: true },
      ],
    });

    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(
        runtimeClient.instanceId,
        undefined,
      ),
    });
  };

  const handleViewLogsClick = (name: string) => {
    const basePath = $page.url.pathname.replace(/\/resources\/?$/, "");
    void goto(`${basePath}/logs?q=${encodeURIComponent(name)}`);
  };

  const handleRefresh = async () => {
    if (dialogResourceKind === ResourceKind.Model) {
      await $createTrigger.mutateAsync({
        models: [
          {
            model: dialogResourceName,
            full: dialogRefreshType === "full",
          },
        ],
      });
    } else {
      await $createTrigger.mutateAsync({
        resources: [{ kind: dialogResourceKind, name: dialogResourceName }],
      });
    }

    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(
        runtimeClient.instanceId,
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
        const status = row.original.meta?.reconcileStatus;
        const isRowReconciling =
          status === V1ReconcileStatus.RECONCILE_STATUS_PENDING ||
          status === V1ReconcileStatus.RECONCILE_STATUS_RUNNING;
        const resourceKey = `${row.original.meta.name.kind}:${row.original.meta.name.name}`;
        return flexRender(ActionsCell, {
          resourceKind: row.original.meta.name.kind,
          resourceName: row.original.meta.name.name,
          resource: row.original,
          canRefresh:
            !isRowReconciling &&
            (row.original.meta.name.kind === ResourceKind.Model ||
              row.original.meta.name.kind === ResourceKind.Source),
          onClickRefreshDialog: openRefreshDialog,
          onClickRefreshErroredPartitions: openRefreshErroredPartitionsDialog,
          onClickViewSpec: openSpecDialog,
          onViewLogsClick: handleViewLogsClick,
          isDropdownOpen: isDropdownOpen(resourceKey),
          onDropdownOpenChange: (isOpen: boolean) =>
            setDropdownOpen(resourceKey, isOpen),
        });
      },
      enableSorting: false,
      meta: {
        widthPercent: 0,
      },
    },
  ];

  $: tableData = data;
</script>

<VirtualizedTable
  data={tableData}
  {columns}
  columnLayout="minmax(95px, 108px) minmax(100px, 3fr) 48px minmax(80px, 2fr) minmax(100px, 2fr) 56px"
  containerHeight={550}
  emptyText="No resources match the current filters"
/>

<RefreshResourceConfirmDialog
  bind:open={isConfirmDialogOpen}
  name={dialogResourceName}
  refreshType={dialogRefreshType}
  onRefresh={handleRefresh}
/>

<RefreshErroredPartitionsDialog
  bind:open={isErroredPartitionsDialogOpen}
  modelName={erroredPartitionsModelName}
  onRefresh={handleRefreshErroredPartitions}
/>

<ResourceSpecDialog
  bind:open={isSpecDialogOpen}
  resourceName={specResourceName}
  resourceKind={specResourceKind}
  resource={specResource}
/>
