<script lang="ts">
  import { page } from "$app/stores";
  import { dynamicHeight } from "@rilldata/web-common/layout/layout-settings.ts";
  import { unorderedParamsAreEqual } from "@rilldata/web-common/lib/url-utils.ts";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onDestroy, onMount } from "svelte";
  import { get } from "svelte/store";
  import { updateThemeVariables } from "../themes/actions";
  import { themeControl } from "../themes/theme-control";
  import CanvasFilters from "./filters/CanvasFilters.svelte";
  import { getCanvasStore } from "./state-managers/state-managers";

  export let maxWidth: number;
  export let clientWidth = 0;
  export let showGrabCursor = false;
  export let filtersEnabled: boolean | undefined;
  export let canvasName: string;
  export let embedded: boolean = false;
  export let homeBookmarkUrlSearch: string | undefined = undefined;
  export let onClick: () => void = () => {};

  let themeBoundary: HTMLElement | null = null;

  let contentRect = new DOMRectReadOnly(0, 0, 0, 0);

  $: ({ instanceId } = $runtime);

  $: ({
    canvasEntity: {
      saveSnapshot,
      restoreSnapshot,
      themeSpec,
      defaultUrlParamsStore,
    },
  } = getCanvasStore(canvasName, instanceId));

  $: ({ width: clientWidth } = contentRect);

  $: if (themeBoundary) {
    $themeControl;
    updateThemeVariables($themeSpec, themeBoundary);
  }

  onMount(async () => {
    await waitUntil(() => !get(defaultUrlParamsStore).isPending, 500);
    const shouldLoadHomeBookmark = unorderedParamsAreEqual(
      $page.url.searchParams,
      get(defaultUrlParamsStore).data,
    );
    await restoreSnapshot(
      shouldLoadHomeBookmark ? homeBookmarkUrlSearch : undefined,
    );
  });

  onDestroy(() => {
    saveSnapshot($page.url.searchParams.toString());
  });
</script>

<main
  class="flex flex-col dashboard-theme-boundary overflow-hidden"
  bind:this={themeBoundary}
  class:w-full={$dynamicHeight}
  class:size-full={!$dynamicHeight}
>
  {#if filtersEnabled}
    <header
      role="presentation"
      class="bg-background border-b py-4 px-2 w-full h-fit select-none z-50 flex items-center justify-center"
      on:click|self={onClick}
    >
      <CanvasFilters {canvasName} {maxWidth} />
    </header>
  {/if}

  <div
    role="presentation"
    id="canvas-scroll-container"
    class="p-2 flex flex-col items-center bg-surface select-none overflow-y-auto overflow-x-hidden"
    class:!cursor-grabbing={showGrabCursor}
    class:w-full={$dynamicHeight}
    class:size-full={!$dynamicHeight}
    class:pb-48={!embedded}
    on:click|self={onClick}
  >
    <div
      class="w-full h-fit flex flex-col items-center row-container relative"
      style:max-width="{maxWidth}px"
      style:min-width="420px"
      bind:contentRect
    >
      <slot />
    </div>
  </div>
</main>

<style>
  div {
    container-type: inline-size;
    container-name: canvas-container;
  }
</style>