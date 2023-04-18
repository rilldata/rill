<script lang="ts">
  import { page } from "$app/stores";
  import ProjectGithubConnection from "../../../../../components/projects/ProjectGithubConnection.svelte";
  import ProjectLogs from "../../../../../components/projects/ProjectLogs.svelte";
  import ProjectStatus from "../../../../../components/projects/ProjectStatus.svelte";
  import ShareProjectCta from "../../../../../components/projects/ShareProjectCTA.svelte";
</script>

<svelte:head>
  <title>Project status</title>
</svelte:head>

<div class="flex flex-col items-center">
  <div class="flex space-x-10 border-b border-gray-200 w-full px-12 py-5">
    <ProjectStatus />
    <ProjectGithubConnection />
    <ShareProjectCta />
  </div>
  <ProjectLogs
    organization={$page.params.organization}
    project={$page.params.project}
  />
</div>
