<script lang="ts">
  import { page } from "$app/stores";
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

<div class="flex flex-col items-center">
  <div class="flex space-x-10 border-b border-gray-200 w-full px-12 py-5">
    <ProjectDeploymentStatus {organization} {project} />
    <ProjectGithubConnection {organization} {project} />
    <ShareProjectCta {organization} {project} />
  </div>
  <ProjectDeploymentLogs {organization} {project} />
</div>
