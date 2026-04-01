<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import {
    createAdminServiceListOrganizationProjectsWithHealth,
    V1DeploymentStatus,
    type V1ProjectHealth,
  } from "@rilldata/web-admin/client";
  import {
    getStatusDotClass,
    getStatusLabel,
  } from "@rilldata/web-admin/features/projects/status/display-utils";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import NameCell from "@rilldata/web-common/features/projects/status/NameCell.svelte";
  import RefreshCell from "@rilldata/web-common/features/projects/status/RefreshCell.svelte";
  import ResourceErrorMessage from "@rilldata/web-common/features/projects/status/ResourceErrorMessage.svelte";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
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

  $: organization = $page.params.organization;

  $: healthQuery = createAdminServiceListOrganizationProjectsWithHealth(
    organization,
    { pageSize: 50 },
  );

  $: allProjects = $healthQuery.data?.projects ?? [];

  function isHealthy(p: V1ProjectHealth): boolean {
    return (
      p.deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING &&
      (p.parseErrorCount ?? 0) === 0 &&
      (p.reconcileErrorCount ?? 0) === 0
    );
  }

  function hasErrors(p: V1ProjectHealth): boolean {
    return (
      p.deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED ||
      (p.parseErrorCount ?? 0) > 0 ||
      (p.reconcileErrorCount ?? 0) > 0
    );
  }

  type StatusFilter = { label: string; value: string };
  const statusFilters: StatusFilter[] = [
    { label: "Healthy", value: "healthy" },
    { label: "Erroring", value: "erroring" },
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
    selectedStatuses = parseArrayParam(
      $page.url.searchParams.get("status"),
    );
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

  $: hasActiveFilters =
    selectedStatuses.length > 0 || searchText.length > 0;

  $: filteredProjects = allProjects.filter((p) => {
    if (selectedStatuses.includes("healthy") && !isHealthy(p)) return false;
    if (selectedStatuses.includes("erroring") && !hasErrors(p)) return false;
    if (
      searchText &&
      !(p.projectName ?? "").toLowerCase().includes(searchText.toLowerCase())
    )
      return false;
    return true;
  });

  function projectReconcileStatus(
    p: V1ProjectHealth,
  ): V1ReconcileStatus {
    if (hasErrors(p)) return V1ReconcileStatus.RECONCILE_STATUS_IDLE;
    return V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  }

  function projectErrorMessage(p: V1ProjectHealth): string {
    const errors: string[] = [];
    if (
      p.deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED
    ) {
      errors.push(
        p.deploymentStatusMessage ?? "Deployment error",
      );
    }
    if ((p.parseErrorCount ?? 0) > 0)
      errors.push(`${p.parseErrorCount} parse error(s)`);
    if ((p.reconcileErrorCount ?? 0) > 0)
      errors.push(`${p.reconcileErrorCount} reconcile error(s)`);
    return errors.join("; ");
  }

  let openDropdownProject = "";
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

  {#if $healthQuery.isLoading}
    <p class="text-sm text-fg-secondary">Loading projects...</p>
  {:else if $healthQuery.isError}
    <p class="text-red-500 text-sm">Failed to load projects</p>
  {:else if filteredProjects.length === 0}
    <p class="text-sm text-fg-secondary py-8 text-center">
      No projects match the current filters
    </p>
  {:else}
    <div class="overflow-x-auto border border-border rounded-sm">
      <table class="w-full text-sm table-fixed">
        <thead>
          <tr class="border-b border-border bg-surface-subtle">
            <th class="px-3 py-2 text-left font-medium text-fg-secondary text-xs">Name</th>
            <th class="px-3 py-2 text-center font-medium text-fg-secondary text-xs w-[48px]">Status</th>
            <th class="px-3 py-2 text-left font-medium text-fg-secondary text-xs w-[160px]">Errors</th>
            <th class="px-3 py-2 text-left font-medium text-fg-secondary text-xs w-[120px]">Last Updated</th>
            <th class="px-3 py-2 w-[56px]"></th>
          </tr>
        </thead>
        <tbody>
          {#each filteredProjects as project (project.projectId)}
            <tr class="border-b border-border last:border-b-0">
              <td class="px-3 py-3 truncate">
                <NameCell name={project.projectName ?? ""} />
              </td>
              <td class="px-3 py-3">
                <ResourceErrorMessage
                  message={projectErrorMessage(project)}
                  status={projectReconcileStatus(project)}
                />
              </td>
              <td class="px-3 py-3">
                {@const errTotal = (project.parseErrorCount ?? 0) + (project.reconcileErrorCount ?? 0)}
                {#if errTotal > 0}
                  <span class="text-red-600 font-medium text-xs">{errTotal}</span>
                  <span class="text-fg-tertiary text-xs ml-1">({project.parseErrorCount ?? 0} parse, {project.reconcileErrorCount ?? 0} reconcile)</span>
                {:else}
                  <span class="text-fg-tertiary">—</span>
                {/if}
              </td>
              <td class="px-3 py-3">
                <RefreshCell date={project.updatedOn ?? ""} />
              </td>
              <td class="px-3 py-3">
                <DropdownMenu.Root
                  open={openDropdownProject === project.projectId}
                  onOpenChange={(isOpen) => {
                    openDropdownProject = isOpen ? (project.projectId ?? "") : "";
                  }}
                >
                  <DropdownMenu.Trigger
                    class="flex-none"
                    aria-label="Project actions"
                  >
                    <IconButton
                      rounded
                      active={openDropdownProject === project.projectId}
                      size={20}
                    >
                      <ThreeDot size="16px" />
                    </IconButton>
                  </DropdownMenu.Trigger>
                  <DropdownMenu.Content align="start">
                    <DropdownMenu.Item
                      class="font-normal flex items-center"
                      href="/{organization}/{project.projectName}/-/status"
                    >
                      <div class="flex items-center">
                        <ExternalLinkIcon size="12px" />
                        <span class="ml-2">View project status</span>
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
