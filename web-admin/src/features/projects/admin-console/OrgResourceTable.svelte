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
    searchText = parseStringParam($page.url.searchParams.get("q"));
  }

  // Sync filter state → URL
  $: if (mounted) {
    filterSync.syncToUrl({
      project: selectedProjects,
      kind: selectedTypes,
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

  $: filteredResources = resources.filter((r) => {
    if (
      selectedProjects.length > 0 &&
      !selectedProjects.includes(r.projectName)
    )
      return false;
    if (selectedTypes.length > 0 && !selectedTypes.includes(r.kind))
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
    searchText = "";
  }

  $: hasActiveFilters =
    selectedProjects.length > 0 ||
    selectedTypes.length > 0 ||
    searchText.length > 0;

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

    {#if hasActiveFilters}
      <button
        class="shrink-0 text-sm text-primary-500 hover:text-primary-600 whitespace-nowrap"
        on:click={clearFilters}
      >
        Clear
      </button>
    {/if}
  </div>

  {#if filteredResources.length === 0}
    <p class="text-fg-secondary text-sm py-8 text-center">
      No resources match the current filters
    </p>
  {:else}
    <div class="overflow-x-auto border border-border rounded-sm">
      <table class="w-full text-sm table-fixed">
        <thead>
          <tr class="border-b border-border bg-surface-subtle">
            <th class="px-3 py-2 text-left font-medium text-fg-secondary text-xs w-[108px]">Type</th>
            <th class="px-3 py-2 text-left font-medium text-fg-secondary text-xs">Name</th>
            <th class="px-3 py-2 text-left font-medium text-fg-secondary text-xs w-[140px]">Project</th>
            <th class="px-3 py-2 text-center font-medium text-fg-secondary text-xs w-[48px]">Status</th>
            <th class="px-3 py-2 text-left font-medium text-fg-secondary text-xs w-[120px]">Last refresh</th>
            <th class="px-3 py-2 w-[56px]"></th>
          </tr>
        </thead>
        <tbody>
          {#each filteredResources as resource (`${resource.projectName}:${resource.kind}:${resource.name}`)}
            {@const resourceKey = `${resource.projectName}:${resource.kind}:${resource.name}`}
            <tr class="border-b border-border last:border-b-0">
              <td class="px-3 py-3">
                <ResourceTypeBadge kind={resource.kind} />
              </td>
              <td class="px-3 py-3 truncate">
                <NameCell name={resource.name} />
              </td>
              <td class="px-3 py-3 text-fg-secondary truncate text-xs">
                {resource.projectName}
              </td>
              <td class="px-3 py-3">
                <ResourceErrorMessage
                  message={resource.reconcileError}
                  status={resource.reconcileError
                    ? V1ReconcileStatus.RECONCILE_STATUS_IDLE
                    : mapReconcileStatus(resource.reconcileStatus)}
                />
              </td>
              <td class="px-3 py-3">
                <RefreshCell date={resource.stateUpdatedOn} />
              </td>
              <td class="px-3 py-3">
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
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>
