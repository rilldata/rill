<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import GithubConnectionDialog from "@rilldata/web-admin/features/projects/github/GithubConnectionDialog.svelte";
  import ProjectGithubConnection from "@rilldata/web-admin/features/projects/github/ProjectGithubConnection.svelte";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import DangerZone from "@rilldata/web-admin/components/settings/DangerZone.svelte";
  import DeleteProject from "@rilldata/web-admin/features/projects/settings/DeleteProject.svelte";
  import HibernateProject from "@rilldata/web-admin/features/projects/settings/HibernateProject.svelte";
  import ProjectNameSettings from "@rilldata/web-admin/features/projects/settings/ProjectNameSettings.svelte";
  import ProjectVisibilitySettings from "@rilldata/web-admin/features/projects/settings/ProjectVisibilitySettings.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: proj = createAdminServiceGetProject(organization, project);
  $: isGithubConnected =
    !!$proj.data?.project?.gitRemote && !$proj.data?.project?.managedGitId;
</script>

<ProjectNameSettings {organization} {project} />

<SettingsContainer title="GitHub" suppressFooter={isGithubConnected}>
  <div slot="body">
    <ProjectGithubConnection {organization} {project} />
  </div>
  <div slot="action">
    <GithubConnectionDialog {organization} {project} />
  </div>
</SettingsContainer>

<div class="danger-zone-section">
  <h3 class="danger-zone-title">Danger Zone</h3>
  <DangerZone>
    <ProjectVisibilitySettings {organization} {project} />
    <HibernateProject {organization} {project} />
    <DeleteProject {organization} {project} />
  </DangerZone>
</div>

<style lang="postcss">
  .danger-zone-section {
    @apply flex flex-col gap-3;
  }

  .danger-zone-title {
    @apply text-lg font-semibold text-red-600;
  }
</style>
