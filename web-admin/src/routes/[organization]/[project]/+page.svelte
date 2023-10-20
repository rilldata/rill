<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import VerticalScrollContainer from "@rilldata/web-common/layout/VerticalScrollContainer.svelte";
  import { createAdminServiceGetProject } from "../../../client";
  import DashboardsTable from "../../../features/dashboards/listing/DashboardsTable.svelte";
  import ProjectHero from "../../../features/projects/ProjectHero.svelte";
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
  <ContentContainer>
    <div class="flex flex-col gap-y-4 items-start w-full">
      <ProjectHero {organization} {project} />
      {#if isProjectDeployed}
        <DashboardsTable {organization} {project} />
      {:else if isProjectHibernating}
        <RedeployProjectCta {organization} {project} />
      {/if}
    </div>
  </ContentContainer>
</VerticalScrollContainer>
