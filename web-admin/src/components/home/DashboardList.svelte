<script lang="ts">
  import { getDashboardsForProject } from "@rilldata/web-admin/components/projects/dashboards";
  import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";
  import { createAdminServiceGetProject } from "../../client";

  export let organization: string;
  export let project: string;

  let dashboards: V1CatalogEntry[];

  $: proj = createAdminServiceGetProject(organization, project);
  $: if ($proj.isSuccess && $proj.data?.productionDeployment) {
    updateDashboardsForProject();
  }

  async function updateDashboardsForProject() {
    dashboards = await getDashboardsForProject($proj.data);
  }
</script>

{#if dashboards?.length === 0}
  <p class="text-gray-500 text-xs">This project has no dashboards yet.</p>
{:else if dashboards?.length > 0}
  <ol>
    {#each dashboards as dashboard}
      <li class="mb-1">
        <a
          href="/{organization}/{project}/{dashboard.name}"
          class="text-gray-700 hover:underline text-xs font-medium leading-4"
          >{dashboard.metricsView?.label ?? dashboard.name}</a
        >
      </li>
    {/each}
  </ol>
{/if}
