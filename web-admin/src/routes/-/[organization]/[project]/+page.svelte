<script lang="ts">
  import { page } from "$app/stores";
  import { useDashboardNames } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useAdminServiceGetProject } from "../../../../client";

  const proj = useAdminServiceGetProject(
    $page.params.organization,
    $page.params.project
  );

  $: dashboardsQuery = useDashboardNames($runtime.instanceId);
</script>

<svelte:head>
  <title>Projects</title>
</svelte:head>

<section class="flex flex-col justify-center items-center h-3/5">
  {#if $proj.isLoading}
    <span>Loading...</span>
  {:else if $proj.isError}
    <span>Error: {$proj.error}</span>
  {:else if $proj.data && $proj.data.project}
    <h1 class="text-3xl font-medium mb-4">
      Project: {$proj.data.project.name}
    </h1>
    <p class="text-lg"><emph>{$proj.data.project.description}</emph></p>
  {/if}
  <div class="mt-4">
    {#if $dashboardsQuery.data}
      {#each $dashboardsQuery.data as dashboard}
        <a
          href="/-/{$page.params.organization}/{$page.params
            .project}/dashboard/{dashboard}"
          class="text-lg">{dashboard}</a
        >
      {/each}
    {/if}
  </div>
</section>
