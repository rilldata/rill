<script lang="ts">
  import { getProjectErrors } from "@rilldata/web-admin/features/projects/getProjectErrors";
  import type { V1ParseError } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import type { Readable } from "svelte/store";
  import { createAdminServiceGetProject } from "../../client";

  export let organization: string;
  export let project: string;

  const queryClient = useQueryClient();

  $: proj = createAdminServiceGetProject(organization, project);
  $: hasReadAccess = $proj.data?.projectPermissions?.readProdStatus;

  let errors: Readable<Array<V1ParseError>>;
  $: if ($proj.data?.prodDeployment?.runtimeInstanceId && hasReadAccess)
    errors = getProjectErrors(
      queryClient,
      $proj.data?.prodDeployment?.runtimeInstanceId
    );
</script>

{#if $proj.isSuccess}
  <ul class="w-full">
    {#if !hasReadAccess}
      <li class="px-12 py-2 font-semibold text-gray-500 border-b">
        You don't have permission to view project logs
      </li>
    {:else if !errors || $errors.length === 0}
      <li class="px-12 py-2 font-semibold text-gray-500 border-b">
        No logs present
      </li>
    {:else}
      <!-- logs -->
      <li class="px-12 py-2 font-semibold text-gray-800 border-b">
        This project has
        <span class="text-red-600">{$errors.length} </span>
        {$errors.length === 1 ? "error" : "errors"}
      </li>
      {#each $errors as error}
        <li
          class="flex gap-x-5 justify-between py-1 px-12 border-b border-gray-200 bg-red-50 font-mono"
        >
          <span class="text-red-600 break-all">
            {error.message}
          </span>
          {#if error.filePath}
            <span class="text-stone-500 font-semibold shrink-0">
              {error.filePath}
            </span>
          {/if}
        </li>
      {/each}
    {/if}
  </ul>
{/if}
