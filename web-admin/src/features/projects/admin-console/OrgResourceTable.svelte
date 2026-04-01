<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
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
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { ExternalLinkIcon } from "lucide-svelte";
  import {
    createUrlFilterSync,
    parseArrayParam,
    parseStringParam,
  } from "@rilldata/web-common/lib/url-filter-sync";

  export let organization: string;

  type OrgResource = {
    projectName: string;
    kind: string;
    name: string;
    reconcileStatus: string;
    reconcileError: string;
    stateUpdatedOn: string;
  };

  export let resources: OrgResource[];

  const filterSync = createUrlFilterSync([
    { key: "project", type: "array" },
    { key: "kind", type: "array" },
    { key: "status", type: "array" },
    { key: "q", type: "string" },
  ]);
  filterSync.init($page.url);

  let searchText = parseStringParam($page.url.searchParams.get("q"));
  let selectedProjects: string[] = parseArrayParam(
    $page.url.searchParams.get("project"),
  );
  let selectedTypes: string[] = parseArrayParam(
    $page.url.searchParams.get("kind"),
  );
  let selectedStatuses: string[] = parseArrayParam(
    $page.url.searchParams.get("status"),
  );

  type StatusFilter = { label: string; value: string };
  const statusFilters: StatusFilter[] = [
    { label: "Healthy", value: "healthy" },
    { label: "Error", value: "error" },
  ];

  let statusDropdownOpen = false;

  let projectDropdownOpen = false;
  let typeDropdownOpen = false;
  let mounted = false;

  // Sync URL → local state on external navigation
  $: if (mounted && filterSync.hasExternalNavigation($page.url)) {
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

  // Sync filter state → URL
  $: if (mounted) {
    filterSync.syncToUrl({
      project: selectedProjects,
      kind: selectedTypes,
      status: selectedStatuses,
      q: searchText,
    });
  }

  onMount(() => {
    mounted = true;
  });

  $: projectNames = [...new Set(resources.map((r) => r.projectName))].sort();

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

  $: filteredResources = resources.filter((r) => {
    if (
      selectedProjects.length > 0 &&
      !selectedProjects.includes(r.projectName)
    )
      return false;
    if (selectedTypes.length > 0 && !selectedTypes.includes(r.kind))
      return false;
    if (selectedStatuses.includes("healthy") && r.reconcileError)
      return false;
    if (selectedStatuses.includes("error") && !r.reconcileError)
      return false;
    if (
      searchText &&
      !r.name.toLowerCase().includes(searchText.toLowerCase()) &&
      !r.projectName.toLowerCase().includes(searchText.toLowerCase())
    )
      return false;
    return true;
  });

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

  $: hasActiveFilters =
    selectedProjects.length > 0 ||
    selectedTypes.length > 0 ||
    selectedStatuses.length > 0 ||
    searchText.length > 0;

  // Sorting
  type SortKey = "type" | "name" | "project" | "status" | "updated";
  let sortKey: SortKey = "name";
  let sortAsc = true;

  function toggleSort(key: SortKey) {
    if (sortKey === key) {
      sortAsc = !sortAsc;
    } else {
      sortKey = key;
      sortAsc = true;
    }
  }

  $: sortedResources = [...filteredResources].sort((a, b) => {
    const dir = sortAsc ? 1 : -1;
    switch (sortKey) {
      case "type":
        return dir * a.kind.localeCompare(b.kind);
      case "name":
        return dir * a.name.localeCompare(b.name);
      case "project":
        return dir * a.projectName.localeCompare(b.projectName);
      case "status": {
        const sa = a.reconcileError ? 1 : 0;
        const sb = b.reconcileError ? 1 : 0;
        return dir * (sa - sb);
      }
      case "updated":
        return dir * (a.stateUpdatedOn ?? "").localeCompare(b.stateUpdatedOn ?? "");
      default:
        return 0;
    }
  });

  let openDropdownKey = "";
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
            {prettyResourceKind(selectedTypes[0])}, +{selectedTypes.length -
              1} other{selectedTypes.length > 2 ? "s" : ""}
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
            {statusFilters.find((s) => s.value === selectedStatuses[0])
              ?.label}, +{selectedStatuses.length - 1} other
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

  {#if sortedResources.length === 0}
    <div class="table-container">
      <div class="flex flex-col items-center gap-y-1 py-10">
        <span class="text-fg-secondary font-semibold text-sm">
          No resources match the current filters
        </span>
      </div>
    </div>
  {:else}
    <div
      class="table-container"
      style:--grid-template-columns="minmax(95px, 130px) minmax(100px, 3fr) minmax(80px, 2fr) 48px minmax(80px, 2fr) 56px"
    >
      <!-- Header -->
      <div class="row bg-surface-subtle sticky top-0 z-10">
        <button
          class="header-cell pl-4"
          on:click={() => toggleSort("type")}
        >
          <span class="truncate">Type</span>
          {#if sortKey === "type"}
            <ArrowDown flip={sortAsc} size="12px" />
          {/if}
        </button>
        <button
          class="header-cell pl-4"
          on:click={() => toggleSort("name")}
        >
          <span class="truncate">Name</span>
          {#if sortKey === "name"}
            <ArrowDown flip={sortAsc} size="12px" />
          {/if}
        </button>
        <button
          class="header-cell pl-4"
          on:click={() => toggleSort("project")}
        >
          <span class="truncate">Project</span>
          {#if sortKey === "project"}
            <ArrowDown flip={sortAsc} size="12px" />
          {/if}
        </button>
        <button
          class="header-cell pl-1"
          on:click={() => toggleSort("status")}
        >
          <span class="truncate">Status</span>
          {#if sortKey === "status"}
            <ArrowDown flip={sortAsc} size="12px" />
          {/if}
        </button>
        <button
          class="header-cell pl-4"
          on:click={() => toggleSort("updated")}
        >
          <span class="truncate">Last refresh</span>
          {#if sortKey === "updated"}
            <ArrowDown flip={sortAsc} size="12px" />
          {/if}
        </button>
        <div class="pl-4 py-2"></div>
      </div>

      <!-- Rows -->
      {#each sortedResources as resource (`${resource.projectName}:${resource.kind}:${resource.name}`)}
        {@const resourceKey = `${resource.projectName}:${resource.kind}:${resource.name}`}
        <div class="row py-3">
          <div class="pl-4 pr-1 flex items-center truncate">
            <ResourceTypeBadge kind={resource.kind} />
          </div>
          <div class="pl-4 pr-1 flex items-center truncate">
            <NameCell name={resource.name} />
          </div>
          <div class="pl-4 pr-1 flex items-center truncate text-fg-secondary text-xs">
            {resource.projectName}
          </div>
          <div class="pl-1 pr-1 flex items-center truncate">
            <ResourceErrorMessage
              message={resource.reconcileError}
              status={resource.reconcileError
                ? V1ReconcileStatus.RECONCILE_STATUS_IDLE
                : mapReconcileStatus(resource.reconcileStatus)}
            />
          </div>
          <div class="pl-4 pr-1 flex items-center truncate">
            <RefreshCell date={resource.stateUpdatedOn} />
          </div>
          <div class="pl-4 pr-1 flex items-center">
            <DropdownMenu.Root
              open={openDropdownKey === resourceKey}
              onOpenChange={(isOpen) => {
                openDropdownKey = isOpen ? resourceKey : "";
              }}
            >
              <DropdownMenu.Trigger
                class="flex-none"
                aria-label="Resource actions"
              >
                <IconButton
                  rounded
                  active={openDropdownKey === resourceKey}
                  size={20}
                >
                  <ThreeDot size="16px" />
                </IconButton>
              </DropdownMenu.Trigger>
              <DropdownMenu.Content align="start">
                <DropdownMenu.Item
                  class="font-normal flex items-center"
                  href="/{organization}/{resource.projectName}/-/status/resources?q={encodeURIComponent(resource.name)}"
                >
                  <div class="flex items-center">
                    <ExternalLinkIcon size="12px" />
                    <span class="ml-2">View in project</span>
                  </div>
                </DropdownMenu.Item>
              </DropdownMenu.Content>
            </DropdownMenu.Root>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style lang="postcss">
  * {
    @apply border-gray-200;
  }

  .table-container {
    @apply flex flex-col border rounded-sm overflow-hidden;
  }

  .row {
    @apply w-fit min-w-full;
    display: grid;
    grid-template-columns: var(--grid-template-columns);
  }

  .row:not(:last-child) {
    @apply border-b;
  }

  .header-cell {
    @apply py-2 font-semibold text-fg-secondary text-left flex flex-row items-center gap-x-1 truncate text-sm;
  }
</style>
