<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import CheckCircleNew from "@rilldata/web-common/components/icons/CheckCircleNew.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import LeaderboardIcon from "@rilldata/web-common/features/canvas/icons/LeaderboardIcon.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { saveExploreDefaults } from "@rilldata/web-common/features/dashboards/workspace/save-explore-defaults";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let exploreName: string;
  export let fileArtifact: FileArtifact;
  export let saving: boolean;

  let justClickedSaveAsDefault = false;

  const { viewingDefaultsStore, dashboardStore } = getStateManagers();

  $: viewingDefaults = $viewingDefaultsStore;
  $: ({ instanceId } = $runtime);
</script>

<Button
  label="Save as default"
  type={!viewingDefaults ? "secondary" : "ghost"}
  preload={false}
  disabled={viewingDefaults}
  onClick={async () => {
    justClickedSaveAsDefault = true;
    await saveExploreDefaults(fileArtifact, $dashboardStore, instanceId, true);
    setTimeout(() => {
      justClickedSaveAsDefault = false;
    }, 2500);
  }}
>
  {#if saving && justClickedSaveAsDefault}
    <LoadingSpinner size="15px" />
    <div class="flex gaep-x-1 items-center">Saving default state</div>
  {:else if viewingDefaults}
    {#if justClickedSaveAsDefault}
      <CheckCircleNew size="15px" className="fill-green-600" />
      <div class="flex gap-x-1 items-center text-green-600">
        Saved default state
      </div>
    {:else}
      <LeaderboardIcon size="16px" color="currentColor" />
      <div class="flex gap-x-1 items-center">Viewing default state</div>
    {/if}
  {:else}
    <LeaderboardIcon size="16px" color="currentColor" />
    <div class="flex gap-x-1 items-center">Save as default</div>
  {/if}
</Button>
