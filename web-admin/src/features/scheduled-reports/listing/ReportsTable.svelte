<script lang="ts">
  import ResourceList from "@rilldata/web-common/features/resources/ResourceList.svelte";
  import ResourceListEmptyState from "@rilldata/web-common/features/resources/ResourceListEmptyState.svelte";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import { TableToolbar } from "@rilldata/web-common/components/table-toolbar";
  import type {
    FilterGroup,
    SortDirection,
  } from "@rilldata/web-common/components/table-toolbar/types";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { renderComponent, type ColumnDef } from "tanstack-table-8-svelte-5";
  import ReportsTableCompositeCell from "./ReportsTableCompositeCell.svelte";

  export let data: V1Resource[];
  export let organization: string;
  export let project: string;

  let searchText = "";
  let selectedResults: string[] = [];
  let sortDirection: SortDirection = "newest";

  function getDisplayName(r: V1Resource): string {
    return r.report?.spec?.displayName || r.meta?.name?.name || "";
  }

  function getLastRun(r: V1Resource): string {
    const last = r.report?.state?.executionHistory?.[0];
    return last?.finishedOn ?? last?.startedOn ?? last?.reportTime ?? "";
  }

  function getResult(r: V1Resource): "ok" | "error" {
    return r.report?.state?.executionHistory?.[0]?.errorMessage
      ? "error"
      : "ok";
  }

  function matchesSearch(r: V1Resource, q: string): boolean {
    if (!q) return true;
    return getDisplayName(r).toLowerCase().includes(q.toLowerCase());
  }

  $: processedData = (data ?? [])
    .filter(
      (r) =>
        matchesSearch(r, searchText) &&
        (selectedResults.length === 0 ||
          selectedResults.includes(getResult(r))),
    )
    .slice()
    .sort((a, b) => {
      const cmp =
        getLastRun(a) < getLastRun(b)
          ? -1
          : getLastRun(a) > getLastRun(b)
            ? 1
            : 0;
      return sortDirection === "newest" ? -cmp : cmp;
    });

  $: filterGroups = [
    {
      label: "Last run",
      key: "result",
      options: [
        { value: "ok", label: "OK" },
        { value: "error", label: "Error" },
      ],
      selected: selectedResults,
      defaultValue: [],
      multiSelect: true,
    },
  ] satisfies FilterGroup[];

  function handleFilterChange(key: string, value: string) {
    if (key !== "result") return;
    selectedResults = selectedResults.includes(value)
      ? selectedResults.filter((v) => v !== value)
      : [...selectedResults, value];
  }

  function clearFilters() {
    selectedResults = [];
    searchText = "";
  }

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
        renderComponent(ReportsTableCompositeCell, {
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
    //     renderComponent(ReportsTableActionCell, {
    //       title: row.original.name,
    //     }),
    // },
  ];

  const columnVisibility = {
    name: false,
    lastRun: false,
  };
</script>

<ResourceList {columns} data={processedData} {columnVisibility} kind="report">
  <TableToolbar
    slot="toolbar"
    {searchText}
    onSearchChange={(t) => (searchText = t)}
    {filterGroups}
    onFilterChange={handleFilterChange}
    onClearAllFilters={clearFilters}
    {sortDirection}
    onSortToggle={() =>
      (sortDirection = sortDirection === "newest" ? "oldest" : "newest")}
    disabled={(data?.length ?? 0) === 0}
  />
  <ResourceListEmptyState
    slot="empty"
    icon={ReportIcon}
    message="You don't have any reports yet"
  >
    <span slot="action">
      Schedule <a
        href="https://docs.rilldata.com/guide/reports/exports"
        target="_blank"
        rel="noopener noreferrer"
      >
        reports</a
      > from any dashboard
    </span>
  </ResourceListEmptyState>
</ResourceList>
