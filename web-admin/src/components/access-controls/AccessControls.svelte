<script lang="ts">
  import {
    createAdminServiceGetProject,
    V1ProjectPermissions,
  } from "@rilldata/web-admin/client";
  import type { CreateQueryResult } from "@tanstack/svelte-query";

  export let organization: string;
  export let project: string;

  let projectPermissions: CreateQueryResult<V1ProjectPermissions>;
  $: projectPermissions = createAdminServiceGetProject(organization, project, {
    query: {
      select: (data) => {
        console.log(data);
        return data.projectPermissions;
      },
    },
  });
  $: console.log($projectPermissions.data);
</script>

{#if $projectPermissions?.data}
  {#if $projectPermissions.data.manageProject}
    <div>
      <slot name="manage-project" />
    </div>
  {:else if $projectPermissions.data.readProject}
    <div>
      <slot name="read-project" />
    </div>
  {/if}
{/if}
