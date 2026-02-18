<script lang="ts">
  import { page } from "$app/stores";
  import ProjectOverview from "@rilldata/web-admin/features/projects/status/overview/ProjectOverview.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<div class="flex flex-col gap-y-4 size-full">
  <ProjectOverview {organization} {project} />
</div>
