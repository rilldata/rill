<script lang="ts">
  import ResourceListEmptyState from "@rilldata/web-common/features/resources/ResourceListEmptyState.svelte";
  import ResourceTable from "@rilldata/web-common/features/resources/ResourceTable.svelte";
  import NameCell from "@rilldata/web-common/features/resources/cells/NameCell.svelte";
  import RelativeTimeCell from "@rilldata/web-common/features/resources/cells/RelativeTimeCell.svelte";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import {
    applyTableFilters,
    TableToolbar,
  } from "@rilldata/web-common/components/table-toolbar";
  import type {
    FilterGroup,
    SortDirection,
  } from "@rilldata/web-common/components/table-toolbar/types";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import cronstrue from "cronstrue";
  import { renderComponent, type ColumnDef } from "tanstack-table-8-svelte-5";
  import ReportActionsCell from "./ReportActionsCell.svelte";
  import ReportFrequencyCell from "./ReportFrequencyCell.svelte";
  import ReportOwnerCell from "./ReportOwnerCell.svelte";
  import ReportStatusCell from "./ReportStatusCell.svelte";

  export let data: V1Resource[];
  export let organization: string;
  export let project: string;

  let searchText = "";
  let selectedResults: string[] = [];
  let sortDirection: SortDirection = "newest";

  function getDisplayName(r: V1Resource): string {
    return r.report?.spec?.displayName || r.meta?.name?.name || "";
  }

  function getCreatedOn(r: V1Resource): string | undefined {
    return r.meta?.createdOn;
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

  function getFrequency(r: V1Resource): string {
    const cron = r.report?.spec?.refreshSchedule?.cron;
    if (!cron) return "";
    try {
      return cronstrue.toString(cron);
    } catch {
      return cron;
    }
  }

  $: processedData = applyTableFilters({
    data: data ?? [],
    searchText,
    matchesSearch,
    filterPredicates: [
      (r) =>
        selectedResults.length === 0 || selectedResults.includes(getResult(r)),
    ],
    sortDirection,
    getSortKey: getCreatedOn,
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

  function handleFilterChange(key: string, selected: string | string[]) {
    if (key !== "result") return;
    selectedResults = Array.isArray(selected) ? selected : [selected];
  }

  function clearFilters() {
    selectedResults = [];
    searchText = "";
  }

  function getRowHref(row: unknown): string {
    const r = row as V1Resource;
    return `reports/${r.meta?.name?.name ?? ""}`;
  }

  const columns: ColumnDef<V1Resource, string>[] = [
    {
      id: "name",
      header: "Report name",
      accessorFn: (row) =>
        row.report?.spec?.displayName || row.meta?.name?.name || "",
      cell: (info) =>
        renderComponent(NameCell, { name: info.getValue() as string }),
    },
    {
      id: "status",
      header: "Status",
      accessorFn: (row) => getResult(row),
      cell: (info) =>
        renderComponent(ReportStatusCell, { resource: info.row.original }),
      meta: { width: "120px" },
    },
    {
      id: "frequency",
      header: "Frequency",
      accessorFn: getFrequency,
      cell: (info) =>
        renderComponent(ReportFrequencyCell, {
          frequency: info.getValue() as string,
        }),
      meta: { width: "180px" },
    },
    // TODO(#9283): add the Tags column once resource-level tags are exposed
    // by the API. https://github.com/rilldata/rill/issues/9283
    {
      id: "lastRun",
      header: "Last run",
      accessorFn: getLastRun,
      cell: (info) =>
        renderComponent(RelativeTimeCell, {
          value: info.getValue() as string,
        }),
      meta: { width: "140px" },
    },
    {
      id: "owner",
      header: "Owner",
      cell: (info) =>
        renderComponent(ReportOwnerCell, {
          organization,
          project,
          ownerId:
            info.row.original.report?.spec?.annotations?.[
              "admin_owner_user_id"
            ] ?? "",
        }),
      enableSorting: false,
      meta: { width: "180px" },
    },
    {
      id: "actions",
      header: "",
      cell: (info) =>
        renderComponent(ReportActionsCell, {
          organization,
          project,
          id: info.row.original.meta?.name?.name ?? "",
          title:
            info.row.original.report?.spec?.displayName ||
            info.row.original.meta?.name?.name ||
            "",
          isCreatedByCode:
            !info.row.original.report?.spec?.annotations?.[
              "admin_owner_user_id"
            ],
        }),
      enableSorting: false,
      meta: { width: "56px", align: "right" },
    },
  ];
</script>

<ResourceTable
  {columns}
  data={processedData}
  kind="report"
  isFiltered={searchText !== "" || selectedResults.length > 0}
  {getRowHref}
>
  <TableToolbar
    slot="toolbar"
    bind:searchText
    {filterGroups}
    onFilterChange={handleFilterChange}
    onClearAllFilters={clearFilters}
    bind:sortDirection
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
</ResourceTable>
