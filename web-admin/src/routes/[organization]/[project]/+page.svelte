<script lang="ts">
  import { page } from "$app/stores";
  import VerticalScrollContainer from "@rilldata/web-common/layout/VerticalScrollContainer.svelte";
  import { createAdminServiceGetProject } from "../../../client";
  import DashboardList from "../../../features/projects/DashboardList.svelte";
  import RedeployProjectCta from "../../../features/projects/RedeployProjectCTA.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: proj = createAdminServiceGetProject(organization, project);
  $: isProjectDeployed = $proj?.data && $proj.data.prodDeployment;
  $: isProjectHibernating = $proj?.data && !$proj.data.prodDeployment;
</script>

<svelte:head>
  <title>{project} overview - Rill</title>
</svelte:head>
<VerticalScrollContainer>
  <div
    class="mx-8 my-8 sm:my-16 sm:mx-16 lg:mx-32 lg:my-24 2xl:mx-40 flex flex-col gap-y-4"
  >
    {#if isProjectDeployed}
      <span class="text-gray-500 text-base font-normal leading-normal"
        >Check out your dashboards below.</span
      >
      <DashboardList {organization} {project} />
    {:else if isProjectHibernating}
      <RedeployProjectCta {organization} {project} />
    {/if}
  </div>
</VerticalScrollContainer>
