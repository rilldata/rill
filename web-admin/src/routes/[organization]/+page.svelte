<script lang="ts">
  import { page } from "$app/stores";
  import {
    useAdminServiceFindOrganization,
    useAdminServiceFindProjects,
  } from "../../client";

  const org = useAdminServiceFindOrganization($page.params.organization);
  const projs = useAdminServiceFindProjects($page.params.organization);
</script>

<svelte:head>
  <title>Organization</title>
</svelte:head>

<section>
  {#if $org.isLoading || $projs.isLoading}
    <span>Loading...</span>
  {:else if $org.isError || $projs.isError}
    <span>Error: {$org.error || $projs.error}</span>
  {:else if $org.data && $org.data.organization}
    <h1>Org: {$org.data.organization.name}</h1>
    <p><emph>{$org.data.organization.description}</emph></p>
    {#if $projs.data && $projs.data.projects}
      <ul>
        {#each $projs.data.projects as proj}
          <li>
            <a href="/{$org.data.organization.name}/{proj.name}">{proj.name}</a>
          </li>
        {/each}
      </ul>
    {/if}
  {/if}
</section>
