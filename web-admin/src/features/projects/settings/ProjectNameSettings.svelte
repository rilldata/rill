<script lang="ts">
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import ProjectRenameForm, {
    ProjectRenameFormId,
  } from "@rilldata/web-admin/features/projects/settings/ProjectRenameForm.svelte";
  import { goto } from "$app/navigation";

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

<SettingsContainer title="Project">
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
      Save
    </Button>
  {/snippet}
</SettingsContainer>
