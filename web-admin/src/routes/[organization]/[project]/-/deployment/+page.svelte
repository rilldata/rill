<script lang="ts">
  import { page } from "$app/stores";
  import ProjectDeploymentLogs from "../../../../../components/projects/ProjectDeploymentLogs.svelte";
  import ProjectGithubConnection from "../../../../../components/projects/ProjectGithubConnection.svelte";
  import ProjectStatus from "../../../../../components/projects/ProjectStatus.svelte";
  import ShareProjectCta from "../../../../../components/projects/ShareProjectCTA.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<svelte:head>
  <title>Project status</title>
</svelte:head>

<div class="flex flex-col items-center">
  <div class="flex space-x-10 border-b border-gray-200 w-full px-12 py-5">
    <ProjectStatus />
    <ProjectGithubConnection {organization} {project} />
    <ShareProjectCta />
  </div>
  <ProjectDeploymentLogs
    organization={$page.params.organization}
    project={$page.params.project}
  />
</div>
