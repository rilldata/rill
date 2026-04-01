<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceListOrganizationProjectsWithHealth,
    V1DeploymentStatus,
    type V1ProjectHealth,
  } from "@rilldata/web-admin/client";
  import {
    getStatusDotClass,
    getStatusLabel,
  } from "@rilldata/web-admin/features/projects/status/display-utils";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { ExternalLinkIcon } from "lucide-svelte";

  $: organization = $page.params.organization;
  $: statusFilter = $page.url.searchParams.get("status") ?? "";

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

  $: filteredProjects = (() => {
    if (statusFilter === "healthy") return allProjects.filter(isHealthy);
    if (statusFilter === "erroring") return allProjects.filter(hasErrors);
    return allProjects;
  })();

  let openDropdownProject = "";
</script>

<div class="flex flex-col gap-y-4">
  <div class="flex items-center gap-x-2">
    <a
      href="/{organization}/-/console/projects"
      class="chip {!statusFilter ? 'chip-active' : ''}"
    >
      All ({allProjects.length})
    </a>
    <a
      href="/{organization}/-/console/projects?status=healthy"
      class="chip {statusFilter === 'healthy' ? 'chip-active' : ''}"
    >
      <span class="w-2 h-2 rounded-full bg-green-500"></span>
      Healthy ({allProjects.filter(isHealthy).length})
    </a>
    <a
      href="/{organization}/-/console/projects?status=erroring"
      class="chip {statusFilter === 'erroring' ? 'chip-active' : ''}"
    >
      <span class="w-2 h-2 rounded-full bg-red-500"></span>
      Erroring ({allProjects.filter(hasErrors).length})
    </a>
  </div>

  {#if $healthQuery.isLoading}
    <p class="text-sm text-fg-secondary">Loading projects...</p>
  {:else if $healthQuery.isError}
    <p class="text-red-500 text-sm">Failed to load projects</p>
  {:else if filteredProjects.length === 0}
    <p class="text-sm text-fg-secondary py-8 text-center">
      {statusFilter ? "No projects match this filter." : "No projects found."}
    </p>
  {:else}
    <div class="overflow-x-auto border border-border rounded-sm">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-border bg-surface-subtle">
            <th class="px-3 py-2 text-left font-medium text-fg-secondary text-xs">Name</th>
            <th class="px-3 py-2 text-center font-medium text-fg-secondary text-xs w-12">Status</th>
            <th class="px-3 py-2 text-left font-medium text-fg-secondary text-xs">Errors</th>
            <th class="px-3 py-2 text-left font-medium text-fg-secondary text-xs">Last Updated</th>
            <th class="px-3 py-2 w-10"></th>
          </tr>
        </thead>
        <tbody>
          {#each filteredProjects as project (project.projectId)}
            {@const errTotal = (project.parseErrorCount ?? 0) + (project.reconcileErrorCount ?? 0)}
            <tr class="border-b border-border last:border-b-0 hover:bg-surface-hover">
              <td class="px-3 py-3 text-fg-primary font-medium truncate max-w-[200px]">
                {project.projectName}
              </td>
              <td class="px-3 py-3">
                <span class="flex items-center justify-center">
                  <span class="w-2 h-2 rounded-full inline-block {getStatusDotClass(project.deploymentStatus ?? V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED)}"></span>
                </span>
              </td>
              <td class="px-3 py-3">
                {#if errTotal > 0}
                  <span class="text-red-600 font-medium">{errTotal}</span>
                  <span class="text-fg-tertiary text-xs ml-1">({project.parseErrorCount ?? 0} parse, {project.reconcileErrorCount ?? 0} reconcile)</span>
                {:else}
                  <span class="text-fg-tertiary">—</span>
                {/if}
              </td>
              <td class="px-3 py-3 text-fg-secondary text-xs">
                {#if project.updatedOn}
                  {new Date(project.updatedOn).toLocaleDateString("en-US", {
                    month: "short",
                    day: "numeric",
                    hour: "numeric",
                    minute: "2-digit",
                  })}
                {:else}
                  —
                {/if}
              </td>
              <td class="px-3 py-3">
                <DropdownMenu.Root
                  open={openDropdownProject === project.projectId}
                  onOpenChange={(isOpen) => {
                    openDropdownProject = isOpen ? (project.projectId ?? "") : "";
                  }}
                >
                  <DropdownMenu.Trigger class="flex-none" aria-label="Project actions">
                    <IconButton rounded active={openDropdownProject === project.projectId} size={20}>
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

<style lang="postcss">
  .chip {
    @apply flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md border border-border bg-surface-subtle no-underline text-inherit font-medium;
  }
  .chip:hover {
    @apply border-primary-500 text-primary-600;
  }
  .chip-active {
    @apply border-primary-500 bg-primary-50 text-primary-600;
  }
</style>
