<script lang="ts">
  import { page } from "$app/stores";
  import { useAdminServiceFindProject } from "../../../client";

  const proj = useAdminServiceFindProject(
    $page.params.organization,
    $page.params.project
  );
</script>

<svelte:head>
  <title>Projects</title>
</svelte:head>

<section>
  {#if $proj.isLoading}
    <span>Loading...</span>
  {:else if $proj.isError}
    <span>Error: {$proj.error}</span>
  {:else if $proj.data && $proj.data.project}
    <h1>Proj: {$proj.data.project.name}</h1>
    <p><emph>{$proj.data.project.description}</emph></p>
  {/if}
</section>
