<script lang="ts">
  import { createAdminServiceListOrganizations } from "../../client";
  import ProjectList from "./ProjectList.svelte";

  const orgs = createAdminServiceListOrganizations();
</script>

{#if $orgs.data && $orgs.data.organizations}
  <ol>
    {#each $orgs.data.organizations as org}
      <li class="mb-3">
        <h2 class="text-base leading-6 font-light mb-1">{org.name}</h2>
        <div class="py-2 px-1.5">
          <ProjectList organization={org.name} />
        </div>
      </li>
    {/each}
  </ol>
{/if}
