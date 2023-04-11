<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "../../../../../client";
  import Logs from "../../../../../components/deployments/Logs.svelte";
  import Status from "../../../../../components/deployments/Status.svelte";

  const proj = createAdminServiceGetProject(
    $page.params.organization,
    $page.params.project
  );
</script>

<svelte:head>
  <title>Project deployment</title>
</svelte:head>

<div class="flex flex-col items-center mx-auto h-3/5">
  {#if $proj.isLoading}
    <span>Loading...</span>
  {:else if $proj.isError}
    <span>Error: {$proj.error}</span>
  {:else if $proj.data && $proj.data.productionDeployment}
    <Status status={$proj.data.productionDeployment.status} />
    <Logs logs={$proj.data.productionDeployment.logs} />
  {/if}
</div>
