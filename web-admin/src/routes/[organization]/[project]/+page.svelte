<script lang="ts">
  import { page } from "$app/stores";
  import { useFindProject } from "../../../client";

  const proj = useFindProject($page.params.organization, $page.params.project);
</script>

<svelte:head>
  <title>Projects</title>
</svelte:head>

<section>
  {#if $proj.isLoading}
    <span>Loading...</span>
  {:else if $proj.isError}
    <span>Error: {$proj.error}</span>
  {:else if $proj.data}
    <h1>Proj: {$proj.data.name}</h1>
    <p><emph>{$proj.data.description}</emph></p>
  {/if}
</section>
