<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import GithubConnectionDialog from "@rilldata/web-admin/features/projects/github/GithubConnectionDialog.svelte";
  import ProjectGithubConnection from "@rilldata/web-admin/features/projects/github/ProjectGithubConnection.svelte";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: proj = createAdminServiceGetProject(organization, project);
  $: isGithubConnected =
    !!$proj.data?.project?.gitRemote && !$proj.data?.project?.managedGitId;
</script>

<div class="flex flex-col gap-y-4 size-full">
  <SettingsContainer title="GitHub" suppressFooter={isGithubConnected}>
    <div slot="body">
      <ProjectGithubConnection {organization} {project} />
    </div>
    <div slot="action">
      <GithubConnectionDialog {organization} {project} />
    </div>
  </SettingsContainer>
</div>
