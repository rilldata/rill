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
      <li class="mb-1">
        <a
          href="/{organization}/{project}/{dashboardListItem.name}"
          class="text-gray-700 hover:underline text-xs font-medium leading-4 {!dashboardListItem.isValid &&
            'italic'}">{dashboardListItem?.title || dashboardListItem.name}</a
        >
      </li>
    {/each}
  </ol>
{/if}
