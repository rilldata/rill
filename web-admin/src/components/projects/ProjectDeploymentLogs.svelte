<script lang="ts">
  import { createAdminServiceGetProject } from "../../client";

  export let organization: string;
  export let project: string;

  $: proj = createAdminServiceGetProject(organization, project);
  $: hasAdminAccess = $proj.data?.projectPermissions?.manageProject;

  $: errors = parseLogs($proj.data?.prodDeployment?.logs);

  interface Error {
    message: string;
    filePath: string;
  }

  function parseLogs(logs: string): Error[] {
    try {
      return JSON.parse(logs).errors;
    } catch (e) {
      return [];
    }
  }
</script>

{#if $proj.isSuccess && errors}
  <ul class="w-full">
    {#if !hasAdminAccess}
      <li class="px-12 py-2 font-semibold text-gray-500 border-b">
        You don't have permission to view project logs
      </li>
    {:else if errors.length === 0}
      <li class="px-12 py-2 font-semibold text-gray-500 border-b">
        No logs present
      </li>
    {:else}
      <!-- logs -->
      <li class="px-12 py-2 font-semibold text-gray-800 border-b">
        This project has
        <span class="text-red-600">{errors.length} </span>
        {errors.length === 1 ? "error" : "errors"}
      </li>
      {#each errors as error}
        <li
          class="flex gap-x-5 justify-between py-1 px-12 border-b border-gray-200 bg-red-50 font-mono"
        >
          <span class="text-red-600 break-all">
            {error.message}
          </span>
          <span class="text-stone-500 font-semibold shrink-0">
            {error.filePath}
          </span>
        </li>
      {/each}
    {/if}
  </ul>
{/if}
