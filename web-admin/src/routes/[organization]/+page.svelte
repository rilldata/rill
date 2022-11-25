<script lang="ts">
  import { page } from "$app/stores";
  import { useFindOrganization, useFindProjects } from "../../client";

  const org = useFindOrganization($page.params.organization);
  const projs = useFindProjects($page.params.organization);
</script>

<svelte:head>
  <title>Organization</title>
</svelte:head>

<section>
  {#if $org.isLoading || $projs.isLoading}
    <span>Loading...</span>
  {:else if $org.isError || $projs.isError}
    <span>Error: {$org.error || $projs.error}</span>
  {:else if $org.data}
    <h1>Org: {$org.data.name}</h1>
    <p><emph>{$org.data.description}</emph></p>
    {#if $projs.data}
      <ul>
        {#each $projs.data as proj}
          <li><a href="/{$org.data.name}/{proj.name}">{proj.name}</a></li>
        {/each}
      </ul>
    {/if}
  {/if}
</section>
