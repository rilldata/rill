<script lang="ts">
  import {
    DashboardListItem,
    getDashboardsForProject,
  } from "@rilldata/web-admin/components/projects/dashboards";
  import { createAdminServiceGetProject } from "../../client";

  export let organization: string;
  export let project: string;

  let dashboardListItems: DashboardListItem[];

  $: proj = createAdminServiceGetProject(organization, project);
  $: if ($proj.isSuccess && $proj.data?.prodDeployment) {
    updateDashboardsForProject();
  }

  async function updateDashboardsForProject() {
    dashboardListItems = await getDashboardsForProject($proj.data);
  }
</script>

{#if dashboardListItems?.length === 0}
  <p class="text-gray-500 text-xs">This project has no dashboards yet.</p>
{:else if dashboardListItems?.length > 0}
  <ol>
    {#each dashboardListItems as dashboardListItem}
      <li class="mb-1 text-xs font-medium leading-4">
        {#if dashboardListItem.isValid}
          <a
            href="/{organization}/{project}/{dashboardListItem.name}"
            class="text-gray-700 hover:underline"
          >
            {dashboardListItem?.title || dashboardListItem.name}
          </a>
        {:else}
          <span class="text-gray-400"
            >{dashboardListItem?.title || dashboardListItem.name}
          </span>
        {/if}
      </li>
    {/each}
  </ol>
{/if}
