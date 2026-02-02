<script lang="ts">
  import { page } from "$app/stores";
  import ProjectNameSettings from "@rilldata/web-admin/features/projects/settings/ProjectNameSettings.svelte";
  import PublicVisibilitySettings from "@rilldata/web-admin/features/projects/settings/PublicVisibilitySettings.svelte";
  import HibernateProject from "@rilldata/web-admin/features/projects/settings/HibernateProject.svelte";
  import DeleteProject from "@rilldata/web-admin/features/projects/settings/DeleteProject.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<div class="flex flex-col gap-6">
  <ProjectNameSettings {organization} {project} />

  <div class="danger-zone">
    <h3 class="danger-zone-title">Danger zone</h3>
    <div class="danger-zone-items">
      <PublicVisibilitySettings {organization} {project} />
      <HibernateProject {organization} {project} />
      <DeleteProject {organization} {project} />
    </div>
  </div>
</div>

<style lang="postcss">
  .danger-zone {
    @apply flex flex-col gap-4;
  }

  .danger-zone-title {
    @apply text-lg font-semibold text-red-600;
  }

  .danger-zone-items {
    @apply flex flex-col gap-4 border border-red-300 rounded-md p-4;
  }
</style>
