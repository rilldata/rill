<script lang="ts">
  import { createAdminServiceListProjectsForOrganization } from "../../client";
  import DashboardList from "./DashboardList.svelte";
  import DeploymentStatusChip from "./DeploymentStatusChip.svelte";

  export let organization: string;

  $: projs = createAdminServiceListProjectsForOrganization(organization);
</script>

{#if $projs.data && $projs.data.projects?.length === 0}
  <p class="text-gray-500 text-xs">This organization has no projects yet.</p>
{:else if $projs.data && $projs.data.projects?.length > 0}
  <ol>
    {#each $projs.data.projects as proj}
      <li class="ml-2">
        <a
          class="flex max-w-fit items-center gap-x-1 mb-1 hover:underline hover:text-gray-700"
          href="{organization}/{proj.name}"
        >
          <DeploymentStatusChip {organization} project={proj.name} iconOnly />
          <h3 class="text-gray-500 font-semibold" style="font-size: 10px;">
            {proj.name.toUpperCase()}
          </h3>
        </a>

        <div class="ml-4">
          <DashboardList {organization} project={proj.name} />
        </div>
      </li>
    {/each}
  </ol>
{/if}
