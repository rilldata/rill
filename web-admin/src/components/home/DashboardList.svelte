<script lang="ts">
  import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";
  import Axios from "axios";
  import {
    createAdminServiceGetProject,
    V1GetProjectResponse,
  } from "../../client";

  export let organization: string;
  export let project: string;

  let dashboards: V1CatalogEntry[];

  $: proj = createAdminServiceGetProject(organization, project);
  $: if ($proj.isSuccess && $proj.data?.prodDeployment) {
    getDashboardsForProject($proj.data);
  }

  async function getDashboardsForProject(projectData: V1GetProjectResponse) {
    // Hack: in development, the runtime host is actually on port 8081
    const runtimeHost = projectData.prodDeployment.runtimeHost.replace(
      "localhost:9091",
      "localhost:8081"
    );

    const axios = Axios.create({
      baseURL: runtimeHost,
      headers: {
        Authorization: `Bearer ${projectData.jwt}`,
      },
    });

    const { data } = await axios.get(
      `/v1/instances/${projectData.prodDeployment.runtimeInstanceId}/catalog?type=OBJECT_TYPE_METRICS_VIEW`
    );

    dashboards = data.entries;

    return;
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
          >{dashboard.metricsView?.label !== ""
            ? dashboard.metricsView.label
            : dashboard.name}</a
        >
      </li>
    {/each}
  </ol>
{/if}
