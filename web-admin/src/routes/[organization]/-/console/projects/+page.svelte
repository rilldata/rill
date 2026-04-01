<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import {
    createAdminServiceListOrganizationProjectsWithHealth,
    V1DeploymentStatus,
    type V1ProjectHealth,
  } from "@rilldata/web-admin/client";
  import VirtualizedTable from "@rilldata/web-common/components/table/VirtualizedTable.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import NameCell from "@rilldata/web-common/features/projects/status/NameCell.svelte";
  import RefreshCell from "@rilldata/web-common/features/projects/status/RefreshCell.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import {
    createUrlFilterSync,
    parseArrayParam,
    parseStringParam,
  } from "@rilldata/web-common/lib/url-filter-sync";
  import {
    isProjectHealthy,
    hasProjectErrors,
  } from "@rilldata/web-admin/features/projects/admin-console/project-health-utils";
  import ErrorCountCell from "@rilldata/web-admin/features/projects/admin-console/ErrorCountCell.svelte";
  import ProjectStatusCell from "@rilldata/web-admin/features/projects/admin-console/ProjectStatusCell.svelte";
  import ProjectActionsCell from "@rilldata/web-admin/features/projects/admin-console/ProjectActionsCell.svelte";
  import type { ColumnDef } from "tanstack-table-8-svelte-5";
  import { renderComponent } from "tanstack-table-8-svelte-5";

  $: organization = $page.params.organization;

  $: healthQuery = createAdminServiceListOrganizationProjectsWithHealth(
    organization,
    { pageSize: 50 },
  );

  $: allProjects = $healthQuery.data?.projects ?? [];

  type StatusFilter = { label: string; value: string };
  const statusFilters: StatusFilter[] = [
    { label: "Healthy", value: "healthy" },
    { label: "Error", value: "error" },
  ];

  const filterSync = createUrlFilterSync([
    { key: "status", type: "array" },
    { key: "q", type: "string" },
  ]);
  filterSync.init($page.url);

  let searchText = parseStringParam($page.url.searchParams.get("q"));
  let selectedStatuses: string[] = parseArrayParam(
    $page.url.searchParams.get("status"),
  );
  let statusDropdownOpen = false;
  let mounted = false;

  $: if (mounted && filterSync.hasExternalNavigation($page.url)) {
    filterSync.markSynced($page.url);
    selectedStatuses = parseArrayParam($page.url.searchParams.get("status"));
    searchText = parseStringParam($page.url.searchParams.get("q"));
  }

  $: if (mounted) {
    filterSync.syncToUrl({
      status: selectedStatuses,
      q: searchText,
    });
  }

  onMount(() => {
    mounted = true;
  });

  function toggleStatus(status: string) {
    if (selectedStatuses.includes(status)) {
      selectedStatuses = selectedStatuses.filter((s) => s !== status);
    } else {
      selectedStatuses = [...selectedStatuses, status];
    }
  }

  function clearFilters() {
    selectedStatuses = [];
    searchText = "";
  }

  $: hasActiveFilters = selectedStatuses.length > 0 || searchText.length > 0;

  $: filteredProjects = allProjects.filter((p) => {
    if (selectedStatuses.length > 0) {
      const matchesAny =
        (selectedStatuses.includes("healthy") && isProjectHealthy(p)) ||
        (selectedStatuses.includes("error") && hasProjectErrors(p));
      if (!matchesAny) return false;
    }
    if (
      searchText &&
      !(p.projectName ?? "").toLowerCase().includes(searchText.toLowerCase())
    )
      return false;
    return true;
  });

  function projectErrorMessage(p: V1ProjectHealth): string {
    const errors: string[] = [];
    if (p.deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED) {
      errors.push(p.deploymentStatusMessage ?? "Deployment error");
    }
    if ((p.parseErrorCount ?? 0) > 0)
      errors.push(`${p.parseErrorCount} parse error(s)`);
    if ((p.reconcileErrorCount ?? 0) > 0)
      errors.push(`${p.reconcileErrorCount} reconcile error(s)`);
    return errors.join("; ");
  }

  let openDropdownProject = "";

  const columns: ColumnDef<V1ProjectHealth, any>[] = [
    {
      accessorKey: "projectName",
      header: "Name",
      cell: ({ getValue }) =>
        renderComponent(NameCell, {
          name: (getValue() as string) ?? "",
        }),
    },
    {
      accessorFn: (row) => projectErrorMessage(row),
      header: "Status",
      sortingFn: (rowA, rowB) => {
        const a = hasProjectErrors(rowA.original)
          ? 2
          : isProjectHealthy(rowA.original)
            ? 0
            : 1;
        const b = hasProjectErrors(rowB.original)
          ? 2
          : isProjectHealthy(rowB.original)
            ? 0
            : 1;
        return a - b;
      },
      cell: ({ row }) =>
        renderComponent(ProjectStatusCell, {
          message: projectErrorMessage(row.original),
        }),
      meta: {
        marginLeft: "1",
      },
    },
    {
      accessorFn: (row) => row.parseErrorCount ?? 0,
      header: "Parse",
      cell: ({ getValue }) =>
        renderComponent(ErrorCountCell, {
          count: getValue() as number,
        }),
    },
    {
      accessorFn: (row) => row.reconcileErrorCount ?? 0,
      header: "Reconcile",
      cell: ({ getValue }) =>
        renderComponent(ErrorCountCell, {
          count: getValue() as number,
        }),
    },
    {
      accessorKey: "updatedOn",
      header: "Last Updated",
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
        const projectId = row.original.projectId ?? "";
        return renderComponent(ProjectActionsCell, {
          href: `/${organization}/${row.original.projectName}/-/status`,
          isDropdownOpen: openDropdownProject === projectId,
          onDropdownOpenChange: (isOpen: boolean) => {
            openDropdownProject = isOpen ? projectId : "";
          },
        });
      },
      enableSorting: false,
      meta: {
        widthPercent: 0,
      },
    },
  ];

  $: tableData = filteredProjects;
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
        on:click={clearFilters}
      >
        Clear
      </button>
    {/if}
  </div>

  {#if $healthQuery.isLoading}
    <p class="text-sm text-fg-secondary">Loading projects...</p>
  {:else if $healthQuery.isError}
    <p class="text-red-500 text-sm">Failed to load projects</p>
  {:else}
    <VirtualizedTable
      data={tableData}
      {columns}
      columnLayout="minmax(100px, 3fr) 48px minmax(60px, 1fr) minmax(80px, 1fr) minmax(100px, 2fr) 56px"
      containerHeight={550}
      emptyText="No projects match the current filters"
    />
  {/if}
</div>
