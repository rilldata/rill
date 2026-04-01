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
  import OverviewCard from "@rilldata/web-common/features/projects/status/overview/OverviewCard.svelte";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { ExternalLinkIcon } from "lucide-svelte";

  $: organization = $page.params.organization;

  $: healthQuery = createAdminServiceListOrganizationProjectsWithHealth(
    organization,
    { pageSize: 50 },
  );

  $: projects = $healthQuery.data?.projects ?? [];

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

  $: totalProjects = projects.length;
  $: healthyCount = projects.filter(isHealthy).length;
  $: errorCount = projects.filter(hasErrors).length;

  let openDropdownProject = "";
</script>

{#if $healthQuery.isLoading}
  <p class="text-sm text-fg-secondary">Loading projects...</p>
{:else if $healthQuery.isError}
  <p class="text-red-500 text-sm">Failed to load projects</p>
{:else}
  <div class="flex flex-col gap-6">
    <OverviewCard title="Summary">
      <div class="chips">
        <div class="chip">
          <span class="font-medium">{totalProjects}</span>
          <span class="text-fg-secondary">{totalProjects === 1 ? "Project" : "Projects"}</span>
        </div>
        <div class="chip chip-green">
          <span class="w-2 h-2 rounded-full bg-green-500"></span>
          <span class="font-medium">{healthyCount}</span>
          <span class="text-fg-secondary">Healthy</span>
        </div>
        <div class="chip chip-red">
          <span class="w-2 h-2 rounded-full bg-red-500"></span>
          <span class="font-medium">{errorCount}</span>
          <span class="text-fg-secondary">Erroring</span>
        </div>
      </div>
    </OverviewCard>

    <OverviewCard title="Projects" viewAllHref="/{organization}/-/console/resources">
      {#if projects.length === 0}
        <p class="text-sm text-fg-secondary">No projects found.</p>
      {:else}
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-border">
                <th class="py-2 text-left font-medium text-fg-secondary text-xs uppercase tracking-wide">Name</th>
                <th class="py-2 text-left font-medium text-fg-secondary text-xs uppercase tracking-wide">Status</th>
                <th class="py-2 text-left font-medium text-fg-secondary text-xs uppercase tracking-wide">Errors</th>
                <th class="py-2 text-left font-medium text-fg-secondary text-xs uppercase tracking-wide">Last Updated</th>
                <th class="py-2 w-10"></th>
              </tr>
            </thead>
            <tbody>
              {#each projects as project (project.projectId)}
                {@const errTotal = (project.parseErrorCount ?? 0) + (project.reconcileErrorCount ?? 0)}
                <tr class="border-b border-border last:border-b-0">
                  <td class="py-3 text-fg-primary font-medium truncate max-w-[200px]">
                    {project.projectName}
                  </td>
                  <td class="py-3">
                    <span class="flex items-center gap-2">
                      <span class="w-2 h-2 rounded-full inline-block {getStatusDotClass(project.deploymentStatus ?? V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED)}"></span>
                      <span class="text-fg-primary">{getStatusLabel(project.deploymentStatus ?? V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED)}</span>
                    </span>
                  </td>
                  <td class="py-3">
                    {#if errTotal > 0}
                      <span class="text-red-600 font-medium">{errTotal}</span>
                      <span class="text-fg-tertiary text-xs ml-1">({project.parseErrorCount ?? 0} parse, {project.reconcileErrorCount ?? 0} reconcile)</span>
                    {:else}
                      <span class="text-fg-tertiary">—</span>
                    {/if}
                  </td>
                  <td class="py-3 text-fg-secondary">
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
                  <td class="py-3">
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
    </OverviewCard>

    <!-- Analytics canvas dashboard placeholder -->
    <OverviewCard title="Analytics">
      <div class="border border-dashed border-border rounded-lg p-8 flex items-center justify-center min-h-[300px]">
        <p class="text-fg-tertiary text-sm">Analytics dashboard coming soon</p>
      </div>
    </OverviewCard>
  </div>
{/if}

<style lang="postcss">
  .chips {
    @apply flex flex-wrap gap-2;
  }
  .chip {
    @apply flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md border border-border bg-surface-subtle;
  }
  .chip-green {
    @apply border-green-200;
  }
  .chip-red {
    @apply border-red-200;
  }
</style>
