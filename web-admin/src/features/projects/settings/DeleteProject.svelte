<script lang="ts">
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import GuardedDeleteProjectConfirmation from "@rilldata/web-admin/features/projects/settings/GuardedDeleteProjectConfirmation.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let { organization, project }: { organization: string; project: string } =
    $props();
</script>

<SettingsContainer title={m.settings_delete_project_title()}>
  {m.settings_delete_project_description()}

  {#snippet action()}
    <GuardedDeleteProjectConfirmation {organization} {project} />
  {/snippet}
</SettingsContainer>
