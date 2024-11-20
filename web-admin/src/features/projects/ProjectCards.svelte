<script lang="ts">
  import { createAdminServiceListProjectsForOrganization } from "../../client";
  import ProjectCard from "./ProjectCard.svelte";

  export let organization: string;

  $: projs = createAdminServiceListProjectsForOrganization(organization);
</script>

<span class="text-gray-500 text-base font-normal leading-normal">
  Check out your projects below.
</span>

{#if $projs.data && $projs.data.projects?.length === 0}
  <p class="text-gray-500 text-xs">This organization has no projects yet.</p>
{:else if $projs.data && $projs.data.projects?.length > 0}
  <ol class="flex gap-6 flex-wrap">
    {#each $projs.data.projects as proj}
      <li>
        <ProjectCard {organization} project={proj.name} />
      </li>
    {/each}
  </ol>
{/if}
