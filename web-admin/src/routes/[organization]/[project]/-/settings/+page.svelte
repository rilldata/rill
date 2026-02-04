<script lang="ts">
  import { page } from "$app/stores";
  import DangerZone from "@rilldata/web-admin/components/settings/DangerZone.svelte";
  import DeleteProject from "@rilldata/web-admin/features/projects/settings/DeleteProject.svelte";
  import ProjectNameSettings from "@rilldata/web-admin/features/projects/settings/ProjectNameSettings.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<ProjectNameSettings {organization} {project} />

<DangerZone>
  <DeleteProject {organization} {project} />
</DangerZone>
