<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { ColumnDef, flexRender } from "@tanstack/svelte-table";
  import Table from "../../../components/table/Table.svelte";
  import { useAlerts } from "../../alerts/selectors";
  import AlertsError from "./AlertsError.svelte";
  import AlertsTableCompositeCell from "./AlertsTableCompositeCell.svelte";
  import AlertsTableEmpty from "./AlertsTableEmpty.svelte";
  import AlertsTableHeader from "./AlertsTableHeader.svelte";
  import NoAlertsCTA from "./NoAlertsCTA.svelte";

  export let organization: string;
  export let project: string;

  $: alerts = useAlerts($runtime.instanceId);
  $: console.log($alerts);

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
          title: info.row.original.alert.spec.title,
          lastTrigger:
            info.row.original.alert.state.executionHistory[0]?.executionTime,
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
    //   id: "nextRun",
    //   accessorFn: (row) => row.nextRun,
    // },
    // {
    //   id: "actions",
    //   cell: ({ row }) =>
    //     flexRender(ReportsTableActionCell, {
    //       title: row.original.name,
    //     }),
    // },
  ];

  const columnVisibility = {
    name: false,
    lastRun: false,
  };
</script>

{#if $alerts.isLoading}
  <div class="m-auto mt-20">
    <Spinner status={EntityStatus.Running} size="24px" />
  </div>
{:else if $alerts.isError}
  <AlertsError />
{:else if $alerts.isSuccess}
  {#if $alerts.data.resources.length === 0}
    <NoAlertsCTA />
  {:else}
    <Table {columns} data={$alerts?.data?.resources} {columnVisibility}>
      <AlertsTableHeader slot="header" />
      <AlertsTableEmpty slot="empty" />
    </Table>
  {/if}
{/if}
