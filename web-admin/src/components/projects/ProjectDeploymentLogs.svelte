<script lang="ts">
  import { createAdminServiceGetProject } from "../../client";

  export let organization: string;
  export let project: string;

  $: proj = createAdminServiceGetProject(organization, project);
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
  {#if errors.length === 0}
    <p class="text-gray-500 my-6">No logs available.</p>
  {:else}
    <!-- logs count -->
    <div class="w-full px-12 py-3 border-b border-gray-200 font-semibold">
      <span class="text-red-600">{errors.length} </span>
      <span class="text-gray-800"> error(s)</span>
    </div>
    <!-- logs -->
    <ul class="w-full">
      {#each errors as error}
        <li
          class="flex justify-between py-1 px-12 border-b border-gray-200 bg-red-50 font-mono"
        >
          <span class="text-red-600">{error.message}</span>
          <span class="text-stone-500 font-semibold">{error.filePath}</span>
        </li>
      {/each}
    </ul>
  {/if}
{/if}
