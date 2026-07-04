<script lang="ts">
  import ResourceListEmptyState from "@rilldata/web-common/features/resources/ResourceListEmptyState.svelte";
  import ResourceTable from "@rilldata/web-common/features/resources/ResourceTable.svelte";
  import NameCell from "@rilldata/web-common/features/resources/cells/NameCell.svelte";
  import RelativeTimeCell from "@rilldata/web-common/features/resources/cells/RelativeTimeCell.svelte";
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
  import AlertActionsCell from "./AlertActionsCell.svelte";
  import AlertOwnerCell from "./AlertOwnerCell.svelte";
  import AlertStatusCell from "./AlertStatusCell.svelte";

  export let data: V1Resource[];
  export let organization: string;
  export let project: string;

  let searchText = "";
  let selectedStatuses: string[] = [];
  let sortDirection: SortDirection = "newest";

  function getDisplayName(r: V1Resource): string {
    return r.alert?.spec?.displayName || r.meta?.name?.name || "";
  }

  function getCreatedOn(r: V1Resource): string | undefined {
    return r.meta?.createdOn;
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
    getSortKey: getCreatedOn,
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

  function getRowHref(row: unknown): string {
    const r = row as V1Resource;
    return `alerts/${r.meta?.name?.name ?? ""}`;
  }

  const columns: ColumnDef<V1Resource, string>[] = [
    {
      id: "name",
      header: "Alert name",
      accessorFn: (row) =>
        row.alert?.spec?.displayName || row.meta?.name?.name || "",
      cell: (info) =>
        renderComponent(NameCell, { name: info.getValue() as string }),
    },
    {
      id: "status",
      header: "Status",
      accessorFn: (row) => getStatus(row),
      cell: (info) =>
        renderComponent(AlertStatusCell, { resource: info.row.original }),
      meta: { width: "120px" },
    },
    // TODO(#9283): add the Tags column once resource-level tags are exposed
    // by the API. https://github.com/rilldata/rill/issues/9283
    {
      id: "lastTriggered",
      header: "Last triggered",
      accessorFn: (row) => {
        const last = row.alert?.state?.executionHistory?.[0];
        return last?.finishedOn ?? last?.startedOn ?? "";
      },
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
        renderComponent(AlertOwnerCell, {
          organization,
          project,
          ownerId:
            info.row.original.alert?.spec?.annotations?.[
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
        renderComponent(AlertActionsCell, {
          organization,
          project,
          id: info.row.original.meta?.name?.name ?? "",
          title:
            info.row.original.alert?.spec?.displayName ||
            info.row.original.meta?.name?.name ||
            "",
          isCreatedByCode:
            !info.row.original.alert?.spec?.annotations?.[
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
  kind="alert"
  isFiltered={searchText !== "" || selectedStatuses.length > 0}
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
</ResourceTable>
