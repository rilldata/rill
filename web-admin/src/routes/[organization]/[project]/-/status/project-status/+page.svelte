<script lang="ts">
  import { page } from "$app/stores";
  import ProjectStatus from "@rilldata/web-admin/features/projects/status/ProjectStatus.svelte";
  import ProjectParseErrors from "@rilldata/web-admin/features/projects/status/ProjectParseErrors.svelte";
  import ProjectResources from "@rilldata/web-admin/features/projects/status/ProjectResources.svelte";
  import ProjectLogs from "@rilldata/web-admin/features/projects/status/ProjectLogs.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<div class="flex flex-col gap-y-4 size-full">
  <ProjectStatus {organization} {project} />
  <ProjectResources />
  <ProjectParseErrors />
  <ProjectLogs />
</div>
