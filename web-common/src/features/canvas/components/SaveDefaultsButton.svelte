<script lang="ts">
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import LeaderboardIcon from "../icons/LeaderboardIcon.svelte";
  import CheckCircleNew from "@rilldata/web-common/components/icons/CheckCircleNew.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";

  export let canvasName: string;
  export let instanceId: string;
  export let saving: boolean;

  let justClickedSaveAsDefault = false;

  $: ({
    canvasEntity: { _rows, setDefaultFilters, _viewingDefaults },
  } = getCanvasStore(canvasName, instanceId));

  $: viewingDefaults = $_viewingDefaults;

  $: rows = $_rows;

  $: canvasIsEmpty = rows.length === 0;
</script>

<Button
  label="Save as default"
  type={!viewingDefaults ? "secondary" : "ghost"}
  preload={false}
  disabled={canvasIsEmpty || viewingDefaults}
  onClick={async () => {
    justClickedSaveAsDefault = true;
    await setDefaultFilters();
    setTimeout(() => {
      justClickedSaveAsDefault = false;
    }, 2500);
  }}
>
  {#if saving && justClickedSaveAsDefault}
    <LoadingSpinner size="15px" />
    <div class="flex gap-x-1 items-center">Saving default filters</div>
  {:else if viewingDefaults}
    {#if justClickedSaveAsDefault}
      <CheckCircleNew size="15px" className="fill-green-600" />
      <div class="flex gap-x-1 items-center text-green-600">
        Saved default filters
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
