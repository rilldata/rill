<script lang="ts">
  import { page } from "$app/stores";
  import ProjectGithubConnection from "@rilldata/web-admin/features/projects/github/ProjectGithubConnection.svelte";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<div class="flex flex-col gap-y-4 size-full">
  <SettingsContainer title="GitHub">
    <div slot="body">
      <ProjectGithubConnection {organization} {project} />
    </div>
  </SettingsContainer>
</div>
