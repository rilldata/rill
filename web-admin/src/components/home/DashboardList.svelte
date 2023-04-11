<script lang="ts">
  import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";
  import Axios from "axios";
  import {
    useAdminServiceGetProject,
    V1GetProjectResponse,
  } from "../../client";

  export let organization: string;
  export let project: string;

  let dashboards: V1CatalogEntry[];

  $: proj = useAdminServiceGetProject(organization, project);
  $: if ($proj.isSuccess && $proj.data?.productionDeployment) {
    getDashboardsForProject($proj.data);
  }

  async function getDashboardsForProject(projectData: V1GetProjectResponse) {
    // Hack: in development, the runtime host is actually on port 8081
    const runtimeHost = projectData.productionDeployment.runtimeHost.replace(
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
      `/v1/instances/${projectData.productionDeployment.runtimeInstanceId}/catalog?type=OBJECT_TYPE_METRICS_VIEW`
    );

    dashboards = data.entries;

    return;
  }
</script>

{#if $proj.isSuccess && dashboards?.length > 0}
  <ol>
    {#each dashboards as dashboard}
      <li class="text-xs text-gray-900 font-medium leading-4 mb-1">
        <a href="/{organization}/{project}/{dashboard.name}">{dashboard.name}</a
        >
      </li>
    {/each}
  </ol>
{/if}
