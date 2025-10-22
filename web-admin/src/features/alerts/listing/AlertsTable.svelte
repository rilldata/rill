<script lang="ts">
  import ResourceHeader from "@rilldata/web-admin/components/table/ResourceHeader.svelte";
  import TableEmptyState from "@rilldata/web-admin/components/table/TableEmptyState.svelte";
  import AlertIcon from "@rilldata/web-common/components/icons/AlertIcon.svelte";
  import { resourceColorMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { flexRender, type ColumnDef } from "@tanstack/svelte-table";
  import Table from "../../../components/table/Table.svelte";
  import AlertsTableCompositeCell from "./AlertsTableCompositeCell.svelte";

  export let data: V1Resource[];
  export let organization: string;
  export let project: string;
  export let showHeader: boolean = false;

  const alertColor = resourceColorMapping[ResourceKind.Alert];

  /**
   * Table column definitions.
   * - "composite": Renders all dashboard data in a single cell.
   * - Others: Used for sorting and filtering but not displayed.
   *
   * Note: TypeScript error prevents using `ColumnDef<DashboardResource, string>[]`.
   * Relevant issues:
   * - https://github.com/TanStack/table/issues/4241
   * - https://github.com/TanStack/table/issues/4302
   */
  const columns: ColumnDef<V1Resource, string>[] = [
    {
      id: "composite",
      cell: (info) =>
        flexRender(AlertsTableCompositeCell, {
          organization: organization,
          project: project,
          id: info.row.original.meta.name.name,
          title:
            info.row.original.alert.spec.displayName ||
            info.row.original.meta.name.name,
          lastTrigger:
            info.row.original.alert.state.executionHistory[0]?.finishedOn ??
            info.row.original.alert.state.executionHistory[0]?.startedOn,
          ownerId:
            info.row.original.alert.spec.annotations["admin_owner_user_id"],
          lastTriggerErrorMessage:
            info.row.original.alert.state.executionHistory[0]?.result
              .errorMessage,
        }),
    },
    {
      id: "name",
      accessorFn: (row) => row.meta.name.name,
    },
    {
      id: "lastRun",
      accessorFn: (row) => row.alert.state.currentExecution?.executionTime,
    },
    // {
    //   id: "actions",
    //   cell: ({ row }) =>
    //     flexRender(AlertsTableActionCell, {
    //       title: row.original.name,
    //     }),
    // },
  ];

  const columnVisibility = {
    name: false,
    lastRun: false,
  };
</script>

<Table {columns} {data} {columnVisibility} kind="alert">
  {#if showHeader}
    <ResourceHeader kind="alert" icon={AlertIcon} slot="header" />
  {/if}
  <TableEmptyState
    slot="empty"
    icon={AlertIcon}
    iconColor={alertColor}
    message="You don't have any alerts yet"
    action="To create an alert, click the Create alert button in a dashboard."
  />
</Table>
