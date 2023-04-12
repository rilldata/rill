<script lang="ts">
  import { createAdminServiceListProjects } from "../../client";
  import DashboardList from "./DashboardList.svelte";

  export let organization: string;

  $: projs = createAdminServiceListProjects(organization);
</script>

{#if $projs.data && $projs.data.projects}
  <ol>
    {#each $projs.data.projects as proj}
      <li class="ml-2">
        <h3 class="text-gray-500 font-semibold mb-1" style="font-size: 10px;">
          {proj.name.toUpperCase()}
        </h3>
        <DashboardList {organization} project={proj.name} />
      </li>
    {/each}
  </ol>
{/if}
