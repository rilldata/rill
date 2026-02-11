<script lang="ts">
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import { getCanvasStore } from "./state-managers/state-managers";
  import StaticCanvasRow from "./StaticCanvasRow.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import Spinner from "../entity-management/Spinner.svelte";
  import { EntityStatus } from "../entity-management/types";
  import { derived } from "svelte/store";
  import {
    getEmbedThemeStoreInstance,
    resolveEmbedTheme,
  } from "../embeds/embed-theme";

  export let canvasName: string;
  export let navigationEnabled: boolean = true;

  $: ({ instanceId } = $runtime);

  $: ({
    canvasEntity: {
      componentsStore,
      _rows,
      firstLoad,
      _maxWidth,
      filtersEnabledStore,
      themeName,
    },
  } = getCanvasStore(canvasName, instanceId));

  $: components = $componentsStore;

  $: filtersEnabled = $filtersEnabledStore;
  $: maxWidth = $_maxWidth;
  $: rows = $_rows;

  const embedThemeStore = getEmbedThemeStoreInstance();
  const embedThemeName = derived([embedThemeStore], () => resolveEmbedTheme());

  // Drive the canvas themeName from the resolved embed theme.
  $: {
    const name = $embedThemeName;
    if (name !== undefined) {
      themeName.set(name ?? undefined);
    }
  }
</script>

{#if canvasName}
  <CanvasDashboardWrapper {maxWidth} {canvasName} {filtersEnabled} embedded>
    {#each rows as row, rowIndex (rowIndex)}
      <StaticCanvasRow
        {row}
        {rowIndex}
        {components}
        {maxWidth}
        {navigationEnabled}
      />
    {:else}
      <div class="size-full flex items-center justify-center">
        {#if $firstLoad}
          <Spinner status={EntityStatus.Running} size="32px" />
        {:else}
          <p class="text-lg text-fg-secondary">No components added</p>
        {/if}
      </div>
    {/each}
  </CanvasDashboardWrapper>
{/if}
