<script lang="ts">
  import ProjectOverview from "@rilldata/web-admin/features/projects/status/overview/ProjectOverview.svelte";
  import { page } from "$app/stores";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<svelte:head>
  <title>{project} status (preview) - Rill</title>
</svelte:head>

<div class="flex flex-col gap-y-4 size-full">
  <ProjectOverview {organization} {project} />
</div>
