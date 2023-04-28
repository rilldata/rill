<script lang="ts">
  import { createAdminServiceListOrganizations } from "../../client";
  import ProjectList from "./ProjectList.svelte";

  const orgs = createAdminServiceListOrganizations();
</script>

{#if $orgs.data && $orgs.data.organizations}
  <ol>
    {#each $orgs.data.organizations as org}
      <li class="mb-2">
        <h2 class="text-base leading-6 font-light mb-1">{org.name}</h2>
        <ProjectList organization={org.name} />
      </li>
    {/each}
  </ol>
{/if}

<style>
  ol:last-child {
    margin-bottom: auto;
  }
</style>
