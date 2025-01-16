<script lang="ts">
  import ResourceHeader from "@rilldata/web-admin/components/table/ResourceHeader.svelte";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { flexRender, type ColumnDef } from "@tanstack/svelte-table";
  import Table from "../../../components/table/Table.svelte";
  import ReportsTableCompositeCell from "./ReportsTableCompositeCell.svelte";

  export let data: V1Resource[];
  export let organization: string;
  export let project: string;

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
        flexRender(ReportsTableCompositeCell, {
          organization,
          project,
          id: info.row.original.meta.name.name,
          title: info.row.original.report.spec.displayName,
          lastRun:
            info.row.original.report.state.executionHistory[0]?.reportTime,
          timeZone: info.row.original.report.spec.refreshSchedule.timeZone,
          frequency: info.row.original.report.spec.refreshSchedule.cron,
          ownerId:
            info.row.original.report.spec.annotations["admin_owner_user_id"],
          lastRunErrorMessage:
            info.row.original.report.state.executionHistory[0]?.errorMessage,
        }),
    },
    {
      id: "name",
      accessorFn: (row) => row.meta.name.name,
    },
    {
      id: "lastRun",
      accessorFn: (row) => row.report.state.currentExecution?.reportTime,
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

<Table {columns} {data} {columnVisibility} kind="report">
  <ResourceHeader kind="report" icon={ReportIcon} slot="header" />
</Table>
