<script lang="ts">
  import ResourceList from "@rilldata/web-common/features/resources/ResourceList.svelte";
  import ResourceListEmptyState from "@rilldata/web-common/features/resources/ResourceListEmptyState.svelte";
  import AlertIcon from "@rilldata/web-common/components/icons/AlertIcon.svelte";
  import {
    applyTableFilters,
    TableToolbar,
  } from "@rilldata/web-common/components/table-toolbar";
  import type {
    FilterGroup,
    SortDirection,
  } from "@rilldata/web-common/components/table-toolbar/types";
  import {
    V1AssertionStatus,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { renderComponent, type ColumnDef } from "tanstack-table-8-svelte-5";
  import AlertsTableCompositeCell from "./AlertsTableCompositeCell.svelte";

  export let data: V1Resource[];
  export let organization: string;
  export let project: string;

  let searchText = "";
  let selectedStatuses: string[] = [];
  let sortDirection: SortDirection = "newest";

  function getDisplayName(r: V1Resource): string {
    return r.alert?.spec?.displayName || r.meta?.name?.name || "";
  }

  function getLastTrigger(r: V1Resource): string {
    const last = r.alert?.state?.executionHistory?.[0];
    return last?.finishedOn ?? last?.startedOn ?? "";
  }

  function getStatus(r: V1Resource): "triggered" | "ok" | "error" | "none" {
    const status = r.alert?.state?.executionHistory?.[0]?.result?.status;
    if (status === V1AssertionStatus.ASSERTION_STATUS_FAIL) return "triggered";
    if (status === V1AssertionStatus.ASSERTION_STATUS_PASS) return "ok";
    if (status === V1AssertionStatus.ASSERTION_STATUS_ERROR) return "error";
    return "none";
  }

  function matchesSearch(r: V1Resource, q: string): boolean {
    if (!q) return true;
    return getDisplayName(r).toLowerCase().includes(q.toLowerCase());
  }

  $: processedData = applyTableFilters({
    data: data ?? [],
    searchText,
    matchesSearch,
    filterPredicates: [
      (r) =>
        selectedStatuses.length === 0 ||
        selectedStatuses.includes(getStatus(r)),
    ],
    sortDirection,
    getSortKey: getLastTrigger,
  });

  $: filterGroups = [
    {
      label: "Status",
      key: "status",
      options: [
        { value: "triggered", label: "Triggered" },
        { value: "ok", label: "OK" },
        { value: "error", label: "Error" },
      ],
      selected: selectedStatuses,
      defaultValue: [],
      multiSelect: true,
    },
  ] satisfies FilterGroup[];

  function handleFilterChange(key: string, selected: string | string[]) {
    if (key !== "status") return;
    selectedStatuses = Array.isArray(selected) ? selected : [selected];
  }

  function clearFilters() {
    selectedStatuses = [];
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
        renderComponent(AlertsTableCompositeCell, {
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
    //     renderComponent(AlertsTableActionCell, {
    //       title: row.original.name,
    //     }),
    // },
  ];

  const columnVisibility = {
    name: false,
    lastRun: false,
  };
</script>

<ResourceList
  {columns}
  data={processedData}
  {columnVisibility}
  kind="alert"
  isFiltered={searchText !== "" || selectedStatuses.length > 0}
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
    icon={AlertIcon}
    message="You don't have any alerts yet"
  >
    <span slot="action">
      Create <a
        href="https://docs.rilldata.com/guide/alerts"
        target="_blank"
        rel="noopener noreferrer"
      >
        alerts
      </a>
      from any dashboard or{" "}
      <a
        href="https://docs.rilldata.com/reference/project-files/alerts"
        target="_blank"
        rel="noopener noreferrer"
      >
        via code</a
      >.
    </span>
  </ResourceListEmptyState>
</ResourceList>
