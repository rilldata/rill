<script lang="ts">
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import { getCanvasStore } from "./state-managers/state-managers";
  import StaticCanvasRow from "./StaticCanvasRow.svelte";
  import CanvasTabGroupView from "./CanvasTabGroupView.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import Spinner from "../entity-management/Spinner.svelte";
  import { EntityStatus } from "../entity-management/types";
  import { page } from "$app/stores";
  import { derived } from "svelte/store";
  import {
    getEmbedThemeStoreInstance,
    resolveEmbedTheme,
  } from "../embeds/embed-theme";

  export let canvasName: string;
  export let navigationEnabled: boolean = true;

  const runtimeClient = useRuntimeClient();

  $: ({ instanceId } = runtimeClient);

  $: ({
    canvasEntity,
    canvasEntity: {
      componentsStore,
      _rows,
      layout,
      firstLoad,
      _maxWidth,
      filtersEnabledStore,
      themeName,
      activeComponent: activeComponentStore,
    },
  } = getCanvasStore(canvasName, instanceId));

  $: components = $componentsStore;
  $: activeComponentId = $activeComponentStore;

  $: filtersEnabled = $filtersEnabledStore;
  $: maxWidth = $_maxWidth;
  $: rows = $_rows;
  $: blocks = $layout;

  // Re-apply the active tab per group whenever the URL changes (e.g. browser back/forward),
  // so deep-linked tab state is restored on navigation, not just on initial load.
  $: if ($page.url.search !== undefined) {
    canvasEntity.applyTabsFromURL();
  }

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
    {#each blocks as block (block.kind === "tab-group" ? `g-${block.group.name}` : `r-${block.rowIndex}`)}
      {#if block.kind === "tab-group"}
        <CanvasTabGroupView
          group={block.group}
          {components}
          {maxWidth}
          {navigationEnabled}
          {activeComponentId}
          onSelect={(tabName) =>
            canvasEntity.setActiveTabInURL(block.group.name, tabName)}
        />
      {:else if rows[block.freeRowIndex]}
        <StaticCanvasRow
          row={rows[block.freeRowIndex]}
          rowIndex={block.rowIndex}
          {components}
          {maxWidth}
          {navigationEnabled}
          {activeComponentId}
        />
      {/if}
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
