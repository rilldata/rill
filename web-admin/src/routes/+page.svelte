<script lang="ts">
  import { useAdminServiceFindOrganizations } from "../client";

  const orgs = useAdminServiceFindOrganizations();
</script>

<svelte:head>
  <title>Organizations</title>
</svelte:head>

<section>
  <h1>Organizations</h1>
  {#if $orgs.isLoading}
    <span>Loading...</span>
  {:else if $orgs.isError}
    <span>Error: {$orgs.error}</span>
  {:else if $orgs.data && $orgs.data.organization}
    <ul>
      {#each $orgs.data.organization as org}
        <li><a href="/{org.name}">{org.name}</a></li>
      {/each}
    </ul>
  {/if}
</section>
