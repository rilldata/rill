<script lang="ts">
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import {
    getCanvasStoreUnguarded,
    type CanvasStore,
  } from "./state-managers/state-managers";
  import StaticCanvasRow from "./StaticCanvasRow.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import Spinner from "../entity-management/Spinner.svelte";
  import { EntityStatus } from "../entity-management/types";
  import { derived, type Readable, type Writable } from "svelte/store";
  import { getEmbedThemeStoreInstance } from "../embeds/embed-theme-store";
  import { resolveEmbedTheme } from "../embeds/embed-theme-utils";

  export let canvasName: string;
  export let navigationEnabled: boolean = true;

  $: ({ instanceId } = $runtime);

  let canvasStore: CanvasStore | undefined;
  let components: CanvasStore["canvasEntity"]["components"];
  let _rows: Readable<unknown>;
  let firstLoad: Readable<boolean>;
  let _maxWidth: Readable<number>;
  let filtersEnabledStore: Readable<boolean | undefined>;
  let themeName: Writable<string | undefined>;

  // Look up the canvas store without throwing if it doesn't exist yet.
  $: canvasStore = getCanvasStoreUnguarded(canvasName, instanceId);

  $: if (canvasStore) {
    ({
      canvasEntity: {
        components,
        _rows,
        firstLoad,
        _maxWidth,
        filtersEnabledStore,
        themeName,
      },
    } = canvasStore);
  }

  $: filtersEnabled = canvasStore ? $filtersEnabledStore : undefined;
  $: maxWidth = canvasStore ? $_maxWidth : 0;
  $: rows = canvasStore ? $_rows : [];

  const embedThemeStore = getEmbedThemeStoreInstance();
  const embedThemeName = derived([embedThemeStore], ([$embedThemeStore]) =>
    resolveEmbedTheme($embedThemeStore),
  );

  $: if (canvasStore) {
    const name = $embedThemeName;
    if (name !== undefined) {
      themeName.set(name ?? undefined);
    }
  }
</script>

{#if canvasName && canvasStore}
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
          <p class="text-lg text-gray-500">No components added</p>
        {/if}
      </div>
    {/each}
  </CanvasDashboardWrapper>
{/if}
