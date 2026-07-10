<script lang="ts">
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import ProjectRenameForm, {
    ProjectRenameFormId,
  } from "@rilldata/web-admin/features/projects/settings/ProjectRenameForm.svelte";
  import { goto } from "$app/navigation";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let { organization, project }: { organization: string; project: string } =
    $props();

  let loading = $state(false);
  let changed = $state(false);

  function onRename(newProject: string) {
    console.log(newProject, project);
    if (newProject !== project) {
      void goto(`/${organization}/${newProject}/-/settings`);
    }
  }
</script>

<SettingsContainer title={m.settings_project_title()}>
  <ProjectRenameForm {organization} {project} bind:loading bind:changed />
  {#snippet action()}
    <Button
      submitForm
      form={ProjectRenameFormId}
      type="primary"
      {loading}
      disabled={!changed}
      {onRename}
    >
      {m.settings_save_button()}
    </Button>
  {/snippet}
</SettingsContainer>
