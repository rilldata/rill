<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "../../../../../client";
  import Logs from "../../../../../components/projects/Logs.svelte";
  import ProjectGithubConnection from "../../../../../components/projects/ProjectGithubConnection.svelte";
  import ProjectStatus from "../../../../../components/projects/ProjectStatus.svelte";
  import ShareProjectCta from "../../../../../components/projects/ShareProjectCTA.svelte";
  import Status from "../../../../../components/projects/Status.svelte";

  const proj = createAdminServiceGetProject(
    $page.params.organization,
    $page.params.project
  );
</script>

<svelte:head>
  <title>Project deployment</title>
</svelte:head>

<div class="flex flex-col items-center">
  <div class="flex space-x-10 border border-black w-full px-12 py-5">
    <ProjectStatus />
    <ProjectGithubConnection />
    <ShareProjectCta />
  </div>
  {#if $proj.isLoading}
    <span>Loading...</span>
  {:else if $proj.isError}
    <span>Error: {$proj.error}</span>
  {:else if $proj.data && $proj.data.productionDeployment}
    <Status status={$proj.data.productionDeployment.status} />
    <Logs logs={$proj.data.productionDeployment.logs} />
  {/if}
</div>
