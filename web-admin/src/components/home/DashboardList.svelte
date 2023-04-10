<script lang="ts">
  import Axios from "axios";
  import {
    useAdminServiceGetProject,
    V1GetProjectResponse,
  } from "../../client";

  export let organization: string;
  export let project: string;

  let dashboards: string[];

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

    const { data: filesData } = await axios.get(
      `/v1/instances/${projectData.productionDeployment.runtimeInstanceId}/files`
    );

    // Filter for dashboard files & extract dashboard names
    dashboards = filesData.paths
      ?.filter((path) => path.includes("dashboards/"))
      // Remove "gitkeep" files
      .filter((path) => !path.includes(".gitkeep"))
      // Remove "dashboards/" prefix and ".yaml" suffix
      .map((path) => path.replace("/dashboards/", "").replace(".yaml", ""))
      // Sort alphabetically case-insensitive
      .sort((a, b) => a.localeCompare(b, undefined, { sensitivity: "base" }));

    // Done
    return;
  }
</script>

{#if $proj.isSuccess && dashboards?.length > 0}
  <ol>
    {#each dashboards as dashboard}
      <li class="text-xs text-gray-900 font-medium leading-4 mb-1">
        <a href="/{organization}/{project}/{dashboard}">{dashboard}</a>
      </li>
    {/each}
  </ol>
{/if}
