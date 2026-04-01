<script lang="ts">
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import {
    getStatusDotClass,
    getStatusLabel,
  } from "@rilldata/web-admin/features/projects/status/display-utils";
  export let organization: string;
  export let projects: Array<{
    name: string;
    status: V1DeploymentStatus;
    updatedOn: string | undefined;
    parseErrorCount: number;
    reconcileErrorCount: number;
  }>;

  function hasResourceErrors(project: {
    parseErrorCount: number;
    reconcileErrorCount: number;
  }): boolean {
    return project.parseErrorCount > 0 || project.reconcileErrorCount > 0;
  }

  function shouldHighlight(project: {
    status: V1DeploymentStatus;
    parseErrorCount: number;
    reconcileErrorCount: number;
  }): boolean {
    return (
      project.status === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED ||
      hasResourceErrors(project)
    );
  }
</script>

{#if projects.length === 0}
  <p class="text-fg-secondary text-sm py-8 text-center">No projects found</p>
{:else}
  <div class="overflow-x-auto rounded-lg border border-border">
    <table class="w-full text-sm">
      <thead>
        <tr class="border-b border-border bg-surface-subtle">
          <th class="px-4 py-3 text-left font-medium text-fg-secondary"
            >Project Name</th
          >
          <th class="px-4 py-3 text-left font-medium text-fg-secondary"
            >Status</th
          >
          <th class="px-4 py-3 text-left font-medium text-fg-secondary"
            >Errors</th
          >
          <th class="px-4 py-3 text-left font-medium text-fg-secondary"
            >Last Updated</th
          >
        </tr>
      </thead>
      <tbody>
        {#each projects as project (project.name)}
          <tr
            class="border-b border-border last:border-b-0 {shouldHighlight(
              project,
            )
              ? 'bg-red-50'
              : ''}"
          >
            <td class="px-4 py-3">
              <a
                href="/{organization}/{project.name}/-/status"
                class="text-primary-500 hover:text-primary-600 font-medium"
              >
                {project.name}
              </a>
            </td>
            <td class="px-4 py-3 text-fg-primary">
              <span class="flex items-center gap-2">
                <span
                  class="inline-block h-2 w-2 rounded-full {getStatusDotClass(
                    project.status,
                  )}"
                ></span>
                {getStatusLabel(project.status)}
              </span>
            </td>
            <td class="px-4 py-3">
              {#if hasResourceErrors(project)}
                <span class="text-red-600 font-medium">
                  {project.parseErrorCount + project.reconcileErrorCount}
                </span>
                <span class="text-fg-tertiary text-xs ml-1">
                  ({project.parseErrorCount} parse, {project.reconcileErrorCount}
                  reconcile)
                </span>
              {:else}
                <span class="text-fg-tertiary">—</span>
              {/if}
            </td>
            <td class="px-4 py-3 text-fg-secondary">
              {#if project.updatedOn}
                {new Date(project.updatedOn).toLocaleDateString("en-US", {
                  month: "short",
                  day: "numeric",
                  year: "numeric",
                  hour: "numeric",
                  minute: "2-digit",
                })}
              {:else}
                —
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}
