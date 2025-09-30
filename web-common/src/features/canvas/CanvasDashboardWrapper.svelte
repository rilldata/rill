<script lang="ts">
  import { exploreName } from "@rilldata/web-common/features/dashboards";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { createSizeChangeHandler } from "@rilldata/web-common/lib/create-size-change-handler.ts";
  import { onDestroy, onMount } from "svelte";
  import CanvasFilters from "./filters/CanvasFilters.svelte";
  import { getCanvasStore } from "./state-managers/state-managers";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { page } from "$app/stores";
  import { updateThemeVariables } from "../themes/actions";

  export let maxWidth: number;
  export let clientWidth = 0;
  export let showGrabCursor = false;
  export let filtersEnabled: boolean | undefined;
  export let canvasName: string;
  export let dynamicHeight: boolean = false;
  export let onClick: () => void = () => {};

  onMount(async () => {
    await restoreSnapshot();
  });

  let contentRect = new DOMRectReadOnly(0, 0, 0, 0);

  $: ({ instanceId } = $runtime);

  $: ({
    canvasEntity: { saveSnapshot, restoreSnapshot, themeSpec },
  } = getCanvasStore(canvasName, instanceId));

  $: ({ width: clientWidth } = contentRect);

  $: updateThemeVariables($themeSpec);

  const sizeChangeHandler = createSizeChangeHandler(
    canvasName,
    ResourceKind.Canvas,
  );

  onDestroy(() => {
    saveSnapshot($page.url.searchParams.toString());
  });
</script>

<main
  class="flex flex-col dashboard-theme-boundary overflow-hidden"
  class:w-full={dynamicHeight}
  class:size-full={!dynamicHeight}
  use:sizeChangeHandler
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
    class="size-full p-2 pb-48 flex flex-col items-center bg-surface select-none overflow-y-auto overflow-x-hidden"
    class:!cursor-grabbing={showGrabCursor}
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
