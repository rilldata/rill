<script lang="ts">
  import Button from "web-common/src/components/button/Button.svelte";
  import CheckCircleNew from "web-common/src/components/icons/CheckCircleNew.svelte";
  import LoadingSpinner from "web-common/src/components/icons/LoadingSpinner.svelte";
  import LeaderboardIcon from "web-common/src/features/canvas/icons/LeaderboardIcon.svelte";
  import { getStateManagers } from "web-common/src/features/dashboards/state-managers/state-managers";
  import { saveExploreDefaults } from "web-common/src/features/dashboards/workspace/save-explore-defaults.ts";
  import type { FileArtifact } from "web-common/src/features/entity-management/file-artifact";
  import { useRuntimeClient } from "web-common/src/runtime-client/v2";
  import { viewingExploreDefaults } from "@rilldata/web-common/features/dashboards/workspace/viewing-explore-defaults.ts";
  import { page } from "$app/state";

  let {
    fileArtifact,
    saving,
  }: {
    fileArtifact: FileArtifact;
    saving: boolean;
  } = $props();

  const runtimeClient = useRuntimeClient();

  let justClickedSaveAsDefault = $state(false);

  const { dashboardStore, validSpecStore, timeRangeSummaryStore } =
    getStateManagers();
  let exploreSpec = $derived($validSpecStore.data?.explore ?? {});
  let metricsViewSpec = $derived($validSpecStore.data?.metricsView ?? {});
  let timeRangeSummary = $derived(
    $timeRangeSummaryStore.data?.timeRangeSummary,
  );

  let viewingDefaults = $derived(
    viewingExploreDefaults(
      page.url.searchParams,
      metricsViewSpec,
      exploreSpec,
      timeRangeSummary,
    ),
  );
</script>

<Button
  label="Save as default"
  type={!viewingDefaults ? "secondary" : "ghost"}
  preload={false}
  disabled={viewingDefaults}
  onClick={async () => {
    justClickedSaveAsDefault = true;
    await saveExploreDefaults(
      runtimeClient,
      fileArtifact,
      $dashboardStore,
      true,
    );
    setTimeout(() => {
      justClickedSaveAsDefault = false;
    }, 2500);
  }}
>
  {#if saving && justClickedSaveAsDefault}
    <LoadingSpinner size="15px" />
    <div class="flex gap-x-1 items-center">Saving default state</div>
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
