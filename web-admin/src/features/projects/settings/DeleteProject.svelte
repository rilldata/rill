<script lang="ts">
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import GuardedDeleteProjectConfirmation from "@rilldata/web-admin/features/projects/settings/GuardedDeleteProjectConfirmation.svelte";

  let { organization, project }: { organization: string; project: string } =
    $props();
</script>

<SettingsContainer title="Delete Project">
  Permanently delete this project and all of its contents from the Rill
  platform. This action is not reversible — please continue with caution.

  {#snippet action()}
    <GuardedDeleteProjectConfirmation {organization} {project} />
  {/snippet}
</SettingsContainer>
