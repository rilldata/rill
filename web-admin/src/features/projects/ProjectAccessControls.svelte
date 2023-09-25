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
      select: (data) => data.projectPermissions,
    },
  });
</script>

{#if $projectPermissions?.data}
  {#if $projectPermissions.data.manageProject}
    <slot name="manage-project" />
  {:else if $projectPermissions.data.readProject}
    <slot name="read-project" />
  {/if}
{/if}
