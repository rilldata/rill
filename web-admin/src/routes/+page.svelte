<script lang="ts">
  import { useFindOrganizations } from "../client";

  const orgs = useFindOrganizations();
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
  {:else if $orgs.data}
    <ul>
      {#each $orgs.data as org}
        <li><a href="/{org.name}">{org.name}</a></li>
      {/each}
    </ul>
  {/if}
</section>
