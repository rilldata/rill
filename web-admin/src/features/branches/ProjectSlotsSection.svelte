<script lang="ts">
  import ProjectSlotsSettings from "@rilldata/web-admin/features/projects/settings/ProjectSlotsSettings.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";

  let { organization, project }: { organization: string; project: string } =
    $props();

  const { cloudEditing } = featureFlags;
</script>

<ProjectSlotsSettings {organization} {project} showDevSlots={$cloudEditing} />
