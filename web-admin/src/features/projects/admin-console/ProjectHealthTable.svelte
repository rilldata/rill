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
  }>;
</script>

{#if projects.length === 0}
  <p class="text-gray-500 text-sm py-8 text-center">No projects found</p>
{:else}
  <div class="overflow-x-auto rounded-lg border border-gray-200">
    <table class="w-full text-sm">
      <thead>
        <tr class="border-b border-gray-200 bg-gray-50">
          <th class="px-4 py-3 text-left font-medium text-gray-600"
            >Project Name</th
          >
          <th class="px-4 py-3 text-left font-medium text-gray-600">Status</th>
          <th class="px-4 py-3 text-left font-medium text-gray-600"
            >Last Updated</th
          >
        </tr>
      </thead>
      <tbody>
        {#each projects as project (project.name)}
          <tr
            class="border-b border-gray-100 last:border-b-0 {project.status ===
            V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED
              ? 'bg-red-50'
              : ''}"
          >
            <td class="px-4 py-3">
              <a
                href="/{organization}/{project.name}/-/status"
                class="text-primary-600 hover:underline font-medium"
              >
                {project.name}
              </a>
            </td>
            <td class="px-4 py-3">
              <span class="flex items-center gap-2">
                <span
                  class="inline-block h-2 w-2 rounded-full {getStatusDotClass(
                    project.status,
                  )}"
                ></span>
                {getStatusLabel(project.status)}
              </span>
            </td>
            <td class="px-4 py-3 text-gray-500">
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
