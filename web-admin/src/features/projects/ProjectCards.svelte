<script lang="ts">
  import { createAdminServiceListProjectsForOrganization } from "../../client";
  import ProjectCard from "./ProjectCard.svelte";

  export let organization: string;

  $: projs = createAdminServiceListProjectsForOrganization(organization);
</script>

{#if $projs.data && $projs.data.projects?.length === 0}
  <p class="text-gray-500 text-xs">This organization has no projects yet.</p>
{:else if $projs.data && $projs.data.projects?.length > 0}
  <ol class="flex gap-6">
    {#each $projs.data.projects as proj}
      <li>
        <ProjectCard {organization} project={proj.name} />
      </li>
    {/each}
  </ol>
{/if}
