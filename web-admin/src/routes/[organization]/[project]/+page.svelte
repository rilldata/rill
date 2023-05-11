<script lang="ts">
  import { page } from "$app/stores";
  import ProjectDashboards from "../../../components/projects/ProjectDashboards.svelte";
  import ProjectDeploymentLogs from "../../../components/projects/ProjectDeploymentLogs.svelte";
  import ProjectDeploymentStatus from "../../../components/projects/ProjectDeploymentStatus.svelte";
  import ProjectGithubConnection from "../../../components/projects/ProjectGithubConnection.svelte";
  import ShareProjectCta from "../../../components/projects/ShareProjectCTA.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<svelte:head>
  <title>{project} overview - Rill</title>
</svelte:head>

<div class="flex flex-col items-center overflow-auto h-full">
  <div
    class="flex flex-row border-b border-gray-200 w-full px-12 py-5 gap-x-10"
  >
    <div class="flex flex-col space-y-4 w-[340px]">
      <ProjectDeploymentStatus {organization} {project} />
      <ProjectGithubConnection {organization} {project} />
    </div>
    <div class="w-[340px]">
      <ProjectDashboards {organization} {project} />
    </div>
    <div class="flex">
      <ShareProjectCta {organization} {project} />
    </div>
  </div>
  <ProjectDeploymentLogs {organization} {project} />
</div>
