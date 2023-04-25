<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { useDashboardNames } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createAdminServiceGetProject } from "../../../client";

  $: proj = createAdminServiceGetProject(
    $page.params.organization,
    $page.params.project
  );

  // Go to first dashboard
  $: dashboardsQuery = useDashboardNames($runtime.instanceId);
  $: if ($dashboardsQuery.data && $dashboardsQuery.data.length > 0) {
    goto(
      `/${$page.params.organization}/${$page.params.project}/${$dashboardsQuery.data[0]}`
    );
  }

  function openDocs() {
    window.open(
      "https://docs.rilldata.com/using-rill/metrics-dashboard",
      "_blank"
    );
  }
</script>

<svelte:head>
  <title>Project</title>
</svelte:head>

<section class="flex flex-col justify-center items-center h-3/5">
  {#if $proj.isLoading || $dashboardsQuery.isLoading}
    <span>Loading...</span>
  {:else if $proj.isError || $dashboardsQuery.isError}
    <span>Error: {$proj.error || $dashboardsQuery.error}</span>
  {:else if $proj.data && $proj.data.project && $dashboardsQuery.data && $dashboardsQuery.data.length === 0}
    <h1 class="text-3xl font-medium text-gray-800 mb-4">
      Project: {$proj.data.project.name}
    </h1>
    <p class="text-lg text-gray-700 mb-6">
      Your project does not have any dashboards... yet!
    </p>
    <Button type="primary" on:click={openDocs}>Read the docs</Button>
  {/if}
</section>
