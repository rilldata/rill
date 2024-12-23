<script lang="ts">
  import ResourceHeader from "@rilldata/web-admin/components/table/ResourceHeader.svelte";
  import ResourceTableEmpty from "@rilldata/web-admin/components/table/ResourceTableEmpty.svelte";
  import Toolbar from "@rilldata/web-admin/components/table/Toolbar.svelte";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { flexRender, type ColumnDef } from "@tanstack/svelte-table";
  import Table from "../../../components/table/Table.svelte";
  import { useReports } from "../selectors";
  import NoReportsCTA from "./NoReportsCTA.svelte";
  import ReportsError from "./ReportsError.svelte";
  import ReportsTableCompositeCell from "./ReportsTableCompositeCell.svelte";

  export let organization: string;
  export let project: string;

  $: ({ instanceId } = $runtime);

  $: reports = useReports(instanceId);

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
          organization: organization,
          project: project,
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

{#if $reports.isLoading}
  <div class="m-auto mt-20">
    <DelayedSpinner isLoading={$reports.isLoading} size="24px" />
  </div>
{:else if $reports.isError}
  <ReportsError />
{:else if $reports.isSuccess}
  {#if $reports.data.resources.length === 0}
    <NoReportsCTA />
  {:else}
    <Table {columns} data={$reports?.data?.resources} {columnVisibility}>
      <Toolbar slot="toolbar" />
      <ResourceHeader kind="report" icon={ReportIcon} slot="header" />
      <ResourceTableEmpty kind="report" slot="empty" />
    </Table>
  {/if}
{/if}
