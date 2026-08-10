<script lang="ts">
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import LeaderboardIcon from "../icons/LeaderboardIcon.svelte";
  import CheckCircleNew from "@rilldata/web-common/components/icons/CheckCircleNew.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
  import { arrayUnorderedEquals } from "@rilldata/web-common/lib/arrayUtils.ts";
  import { page } from "$app/state";

  let {
    canvasName,
    instanceId,
    saving,
  }: {
    canvasName: string;
    instanceId: string;
    saving: boolean;
  } = $props();

  let justClickedSaveAsDefault = $state(false);

  let {
    canvasEntity: {
      _rows,
      saveDefaultFilters,
      defaultUrlParamsStore,
      dashboardProvider,
    },
  } = $derived(getCanvasStore(canvasName, instanceId));

  let defaultUrlParams = $derived($defaultUrlParamsStore);
  let curUrlParams = $derived(page.url.searchParams);

  let viewingDefaults = $derived.by(() => {
    if (
      !arrayUnorderedEquals(
        Object.keys(dashboardProvider.yamlConfigProvider.specPinnedFilters),
        Object.keys(dashboardProvider.yamlConfigProvider.pinnedFilters),
      )
    ) {
      return false;
    }
    if (
      !arrayUnorderedEquals(
        Object.keys(dashboardProvider.yamlConfigProvider.specRequiredFilters),
        Object.keys(dashboardProvider.yamlConfigProvider.requiredFilters),
      )
    ) {
      return false;
    }
    if (defaultUrlParams.size === 0) {
      return false;
    }

    for (const [key, value] of defaultUrlParams.entries()) {
      if (curUrlParams.get(key) !== value) {
        // Ignore time range if not set
        if (
          curUrlParams.get(key) === null &&
          key === ExploreStateURLParams.TimeRange
        ) {
          continue;
        }
        return false;
      }
    }
    for (const [key, value] of curUrlParams.entries()) {
      if (defaultUrlParams.get(key) !== value) {
        return false;
      }
    }
    return true;
  });

  let rows = $derived($_rows);
  let canvasIsEmpty = $derived(rows.length === 0);
</script>

<Button
  label={m.canvas_save_as_default()}
  type={!viewingDefaults ? "secondary" : "ghost"}
  preload={false}
  disabled={canvasIsEmpty || viewingDefaults}
  onClick={async () => {
    justClickedSaveAsDefault = true;
    await saveDefaultFilters();
    setTimeout(() => {
      justClickedSaveAsDefault = false;
    }, 2500);
  }}
>
  {#snippet children()}
    {#if saving && justClickedSaveAsDefault}
      <LoadingSpinner size="15px" />
      <div class="flex gap-x-1 items-center">
        {m.canvas_saving_default_filters()}
      </div>
    {:else if viewingDefaults}
      {#if justClickedSaveAsDefault}
        <CheckCircleNew size="15px" className="fill-green-600" />
        <div class="flex gap-x-1 items-center text-green-600">
          {m.canvas_saved_default_filters()}
        </div>
      {:else}
        <LeaderboardIcon size="16px" color="currentColor" />
        <div class="flex gap-x-1 items-center">
          {m.canvas_viewing_default_state()}
        </div>
      {/if}
    {:else}
      <LeaderboardIcon size="16px" color="currentColor" />
      <div class="flex gap-x-1 items-center">{m.canvas_save_as_default()}</div>
    {/if}
  {/snippet}
</Button>
