<script lang="ts">
  import { page } from "$app/stores";
  import DangerZone from "@rilldata/web-admin/components/settings/DangerZone.svelte";
  import DeleteProject from "@rilldata/web-admin/features/projects/settings/DeleteProject.svelte";
  import HibernateProject from "@rilldata/web-admin/features/projects/settings/HibernateProject.svelte";
  import ProjectNameSettings from "@rilldata/web-admin/features/projects/settings/ProjectNameSettings.svelte";
  import ProjectVisibilitySettings from "@rilldata/web-admin/features/projects/settings/ProjectVisibilitySettings.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<ProjectNameSettings {organization} {project} />

<div class="danger-zone-section">
  <h3 class="danger-zone-title">Danger Zone</h3>
  <DangerZone>
    <ProjectVisibilitySettings {organization} {project} />
  </DangerZone>
  <DangerZone>
    <HibernateProject {organization} {project} />
  </DangerZone>
  <DangerZone>
    <DeleteProject {organization} {project} />
  </DangerZone>
</div>

<style lang="postcss">
  .danger-zone-section {
    @apply flex flex-col gap-4;
  }

  .danger-zone-title {
    @apply text-lg font-semibold text-red-600;
  }
</style>
