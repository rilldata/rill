<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import VirtualizedTable from "@rilldata/web-common/components/table/VirtualizedTable.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import {
    ResourceKind,
    prettyResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import ResourceErrorMessage from "@rilldata/web-common/features/projects/status/ResourceErrorMessage.svelte";
  import NameCell from "@rilldata/web-common/features/projects/status/NameCell.svelte";
  import RefreshCell from "@rilldata/web-common/features/projects/status/RefreshCell.svelte";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import {
    createUrlFilterSync,
    parseArrayParam,
    parseStringParam,
  } from "@rilldata/web-common/lib/url-filter-sync";
  import type { ColumnDef } from "tanstack-table-8-svelte-5";
  import { renderComponent } from "tanstack-table-8-svelte-5";
  import OrgActionsCell from "./OrgActionsCell.svelte";
  import ProjectNameCell from "./ProjectNameCell.svelte";

  type OrgResource = {
    projectName: string;
    kind: string;
    name: string;
    reconcileStatus: string;
    reconcileError: string;
    stateUpdatedOn: string;
  };

  let {
    organization,
    resources,
  }: {
    organization: string;
    resources: OrgResource[];
  } = $props();

  const filterSync = createUrlFilterSync([
    { key: "project", type: "array" },
    { key: "kind", type: "array" },
    { key: "status", type: "array" },
    { key: "q", type: "string" },
  ]);
  filterSync.init($page.url);

  let searchText = $state(
    parseStringParam($page.url.searchParams.get("q")),
  );
  let selectedProjects: string[] = $state(
    parseArrayParam($page.url.searchParams.get("project")),
  );
  let selectedTypes: string[] = $state(
    parseArrayParam($page.url.searchParams.get("kind")),
  );
  let selectedStatuses: string[] = $state(
    parseArrayParam($page.url.searchParams.get("status")),
  );

  type StatusFilter = { label: string; value: string };
  const statusFilters: StatusFilter[] = [
    { label: "Healthy", value: "healthy" },
    { label: "Error", value: "error" },
  ];

  let statusDropdownOpen = $state(false);
  let projectDropdownOpen = $state(false);
  let typeDropdownOpen = $state(false);
  let mounted = $state(false);

  $effect(() => {
    if (mounted && filterSync.hasExternalNavigation($page.url)) {
      filterSync.markSynced($page.url);
      selectedProjects = parseArrayParam(
        $page.url.searchParams.get("project"),
      );
      selectedTypes = parseArrayParam($page.url.searchParams.get("kind"));
      selectedStatuses = parseArrayParam(
        $page.url.searchParams.get("status"),
      );
      searchText = parseStringParam($page.url.searchParams.get("q"));
    }
  });

  $effect(() => {
    if (mounted) {
      filterSync.syncToUrl({
        project: selectedProjects,
        kind: selectedTypes,
        status: selectedStatuses,
        q: searchText,
      });
    }
  });

  onMount(() => {
    mounted = true;
  });

  let projectNames = $derived(
    [...new Set(resources.map((r) => r.projectName))].sort(),
  );

  const filterableTypes = [
    ResourceKind.Source,
    ResourceKind.Model,
    ResourceKind.MetricsView,
    ResourceKind.Explore,
    ResourceKind.Canvas,
    ResourceKind.Theme,
    ResourceKind.Report,
    ResourceKind.Alert,
    ResourceKind.API,
    ResourceKind.Connector,
  ];

  function toggleStatus(status: string) {
    if (selectedStatuses.includes(status)) {
      selectedStatuses = selectedStatuses.filter((s) => s !== status);
    } else {
      selectedStatuses = [...selectedStatuses, status];
    }
  }

  let filteredResources = $derived(
    resources.filter((r) => {
      if (
        selectedProjects.length > 0 &&
        !selectedProjects.includes(r.projectName)
      )
        return false;
      if (selectedTypes.length > 0 && !selectedTypes.includes(r.kind))
        return false;
      if (selectedStatuses.length > 0) {
        const matchesAny =
          (selectedStatuses.includes("healthy") && !r.reconcileError) ||
          (selectedStatuses.includes("error") && !!r.reconcileError);
        if (!matchesAny) return false;
      }
      if (
        searchText &&
        !r.name.toLowerCase().includes(searchText.toLowerCase()) &&
        !r.projectName.toLowerCase().includes(searchText.toLowerCase())
      )
        return false;
      return true;
    }),
  );

  function mapReconcileStatus(status: string): V1ReconcileStatus {
    switch (status) {
      case "RECONCILE_STATUS_IDLE":
        return V1ReconcileStatus.RECONCILE_STATUS_IDLE;
      case "RECONCILE_STATUS_PENDING":
        return V1ReconcileStatus.RECONCILE_STATUS_PENDING;
      case "RECONCILE_STATUS_RUNNING":
        return V1ReconcileStatus.RECONCILE_STATUS_RUNNING;
      default:
        return V1ReconcileStatus.RECONCILE_STATUS_IDLE;
    }
  }

  function toggleProject(project: string) {
    if (selectedProjects.includes(project)) {
      selectedProjects = selectedProjects.filter((p) => p !== project);
    } else {
      selectedProjects = [...selectedProjects, project];
    }
  }

  function toggleType(type: string) {
    if (selectedTypes.includes(type)) {
      selectedTypes = selectedTypes.filter((t) => t !== type);
    } else {
      selectedTypes = [...selectedTypes, type];
    }
  }

  function clearFilters() {
    selectedProjects = [];
    selectedTypes = [];
    selectedStatuses = [];
    searchText = "";
  }

  let hasActiveFilters = $derived(
    selectedProjects.length > 0 ||
      selectedTypes.length > 0 ||
      selectedStatuses.length > 0 ||
      searchText.length > 0,
  );

  let openDropdownKey = $state("");

  const columns: ColumnDef<OrgResource, any>[] = [
    {
      accessorKey: "kind",
      header: "Type",
      cell: ({ row }) =>
        renderComponent(ResourceTypeBadge, {
          kind: row.original.kind as ResourceKind,
        }),
    },
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ getValue }) =>
        renderComponent(NameCell, {
          name: getValue() as string,
        }),
    },
    {
      accessorKey: "projectName",
      header: "Project",
      cell: ({ getValue }) =>
        renderComponent(ProjectNameCell, {
          name: getValue() as string,
        }),
    },
    {
      accessorFn: (row) => row.reconcileError,
      header: "Status",
      sortingFn: (rowA, rowB) => {
        const a = rowA.original.reconcileError ? 1 : 0;
        const b = rowB.original.reconcileError ? 1 : 0;
        return a - b;
      },
      cell: ({ row }) =>
        renderComponent(ResourceErrorMessage, {
          message: row.original.reconcileError,
          status: row.original.reconcileError
            ? V1ReconcileStatus.RECONCILE_STATUS_IDLE
            : mapReconcileStatus(row.original.reconcileStatus),
        }),
      meta: {
        marginLeft: "1",
      },
    },
    {
      accessorKey: "stateUpdatedOn",
      header: "Last refresh",
      sortDescFirst: true,
      cell: (info) =>
        renderComponent(RefreshCell, {
          date: (info.getValue() as string) ?? "",
        }),
    },
    {
      accessorKey: "actions",
      header: "",
      cell: ({ row }) => {
        const resourceKey = `${row.original.projectName}:${row.original.kind}:${row.original.name}`;
        return renderComponent(OrgActionsCell, {
          href: `/${organization}/${row.original.projectName}/-/status/resources?q=${encodeURIComponent(row.original.name)}`,
          isDropdownOpen: openDropdownKey === resourceKey,
          onDropdownOpenChange: (isOpen: boolean) => {
            openDropdownKey = isOpen ? resourceKey : "";
          },
        });
      },
      enableSorting: false,
      meta: {
        widthPercent: 0,
      },
    },
  ];

  let tableData = $derived(filteredResources);
</script>

<div class="flex flex-col gap-y-4">
  <div class="flex flex-row items-center gap-x-4 min-h-9">
    <div class="flex-1 min-w-0 min-h-9">
      <Search
        bind:value={searchText}
        placeholder="Search"
        large
        autofocus={false}
        showBorderOnFocus={false}
        retainValueOnMount
      />
    </div>

    <DropdownMenu.Root bind:open={projectDropdownOpen}>
      <DropdownMenu.Trigger
        class="min-w-fit min-h-9 flex flex-row gap-1 items-center rounded-sm border bg-input {projectDropdownOpen
          ? 'bg-gray-200'
          : 'hover:bg-surface-hover'} px-2 py-1"
      >
        <span class="text-fg-secondary font-medium">
          {#if selectedProjects.length === 0}
            All projects
          {:else if selectedProjects.length === 1}
            {selectedProjects[0]}
          {:else}
            {selectedProjects[0]}, +{selectedProjects.length - 1} other{selectedProjects.length >
            2
              ? "s"
              : ""}
          {/if}
        </span>
        {#if projectDropdownOpen}
          <CaretUpIcon size="12px" />
        {:else}
          <CaretDownIcon size="12px" />
        {/if}
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="start" class="w-48">
        {#each projectNames as project}
          <DropdownMenu.CheckboxItem
            closeOnSelect={false}
            checked={selectedProjects.includes(project)}
            onCheckedChange={() => toggleProject(project)}
          >
            {project}
          </DropdownMenu.CheckboxItem>
        {/each}
      </DropdownMenu.Content>
    </DropdownMenu.Root>

    <DropdownMenu.Root bind:open={typeDropdownOpen}>
      <DropdownMenu.Trigger
        class="min-w-fit min-h-9 flex flex-row gap-1 items-center rounded-sm border bg-input {typeDropdownOpen
          ? 'bg-gray-200'
          : 'hover:bg-surface-hover'} px-2 py-1"
      >
        <span class="text-fg-secondary font-medium">
          {#if selectedTypes.length === 0}
            All types
          {:else if selectedTypes.length === 1}
            {prettyResourceKind(selectedTypes[0])}
          {:else}
            {prettyResourceKind(selectedTypes[0])}, +{selectedTypes.length - 1} other{selectedTypes.length >
            2
              ? "s"
              : ""}
          {/if}
        </span>
        {#if typeDropdownOpen}
          <CaretUpIcon size="12px" />
        {:else}
          <CaretDownIcon size="12px" />
        {/if}
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="start" class="w-48">
        {#each filterableTypes as type}
          <DropdownMenu.CheckboxItem
            closeOnSelect={false}
            checked={selectedTypes.includes(type)}
            onCheckedChange={() => toggleType(type)}
          >
            {prettyResourceKind(type)}
          </DropdownMenu.CheckboxItem>
        {/each}
      </DropdownMenu.Content>
    </DropdownMenu.Root>

    <DropdownMenu.Root bind:open={statusDropdownOpen}>
      <DropdownMenu.Trigger
        class="min-w-fit min-h-9 flex flex-row gap-1 items-center rounded-sm border bg-input {statusDropdownOpen
          ? 'bg-gray-200'
          : 'hover:bg-surface-hover'} px-2 py-1"
      >
        <span class="text-fg-secondary font-medium">
          {#if selectedStatuses.length === 0}
            All statuses
          {:else if selectedStatuses.length === 1}
            {statusFilters.find((s) => s.value === selectedStatuses[0])
              ?.label ?? selectedStatuses[0]}
          {:else}
            {statusFilters.find((s) => s.value === selectedStatuses[0])?.label},
            +{selectedStatuses.length - 1} other
          {/if}
        </span>
        {#if statusDropdownOpen}
          <CaretUpIcon size="12px" />
        {:else}
          <CaretDownIcon size="12px" />
        {/if}
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="start" class="w-48">
        {#each statusFilters as status}
          <DropdownMenu.CheckboxItem
            closeOnSelect={false}
            checked={selectedStatuses.includes(status.value)}
            onCheckedChange={() => toggleStatus(status.value)}
          >
            {status.label}
          </DropdownMenu.CheckboxItem>
        {/each}
      </DropdownMenu.Content>
    </DropdownMenu.Root>

    {#if hasActiveFilters}
      <button
        class="shrink-0 text-sm text-primary-500 hover:text-primary-600 whitespace-nowrap"
        onclick={clearFilters}
      >
        Clear
      </button>
    {/if}
  </div>

  <VirtualizedTable
    data={tableData}
    {columns}
    columnLayout="minmax(95px, 130px) minmax(100px, 3fr) minmax(80px, 2fr) 48px minmax(80px, 2fr) 56px"
    containerHeight={550}
    emptyText="No resources match the current filters"
  />
</div>
