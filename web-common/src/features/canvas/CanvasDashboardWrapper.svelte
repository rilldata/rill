<script lang="ts">
  import { dynamicHeight } from "@rilldata/web-common/layout/layout-settings.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import CellInspector from "@rilldata/web-common/components/CellInspector.svelte";
  import WarningIcon from "@rilldata/web-common/components/icons/WarningIcon.svelte";
  import CanvasFilters from "./filters/CanvasFilters.svelte";
  import { getCanvasStore } from "./state-managers/state-managers";
  import ThemeProvider from "../dashboards/ThemeProvider.svelte";
  import CanvasPdfExportHeader from "../exports/pdf/CanvasPdfExportHeader.svelte";

  const client = useRuntimeClient();

  export let maxWidth: number;
  export let clientWidth = 0;
  export let showGrabCursor = false;
  export let filtersEnabled: boolean | undefined;
  export let canvasName: string;
  export let embedded: boolean = false;
  export let builder = false;
  export let onClick: () => void = () => {};

  let contentRect = new DOMRectReadOnly(0, 0, 0, 0);

  $: ({ instanceId } = client);

  $: ({
    canvasEntity: {
      theme,
      filterManager: { missingRequiredFiltersStore },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: missingRequiredFilters = $missingRequiredFiltersStore;
  $: hasMissingRequired = missingRequiredFilters.length > 0;

  $: ({ width: clientWidth } = contentRect);
</script>

<ThemeProvider theme={$theme}>
  <main
    class="flex flex-col overflow-hidden"
    class:w-full={$dynamicHeight}
    class:size-full={!$dynamicHeight}
  >
    {#if filtersEnabled}
      <header
        role="presentation"
        class="bg-surface-subtle border-b py-4 px-2 w-full h-fit select-none z-50 flex items-center justify-center"
        onclick={(e) => {
          if (e.target === e.currentTarget) onClick();
        }}
      >
        <CanvasFilters {canvasName} {maxWidth} {builder} />
      </header>
    {/if}

    <!-- Off-screen read-only header used only as the PDF capture target. -->
    <div
      aria-hidden="true"
      class="pointer-events-none absolute"
      style="left: -99999px; top: 0;"
    >
      <CanvasPdfExportHeader {canvasName} {instanceId} {maxWidth} />
    </div>

    <div
      role="presentation"
      id="canvas-scroll-container"
      class="p-2 flex flex-col items-center bg-surface-background select-none overflow-y-auto overflow-x-hidden"
      class:!cursor-grabbing={showGrabCursor}
      class:w-full={$dynamicHeight}
      class:size-full={!$dynamicHeight}
      class:pb-48={!embedded}
      onclick={(e) => {
        if (e.target === e.currentTarget) onClick();
      }}
    >
      {#if hasMissingRequired}
        <div class="w-full flex justify-center px-6 pt-24 pb-12">
          <div
            class="flex flex-col items-center text-center gap-y-3 px-8 py-10 rounded-lg border border-gray-200 bg-surface-subtle shadow-sm w-full max-w-lg"
            role="alert"
          >
            <WarningIcon size="32px" className="text-amber-500" />
            <h2 class="text-lg font-semibold text-fg-primary">
              Select a value to continue
            </h2>
            <p class="text-sm text-fg-secondary">
              This dashboard requires values for the following filter{missingRequiredFilters.length >
              1
                ? "s"
                : ""}:
            </p>
            <ul
              class="text-sm text-fg-primary flex flex-wrap justify-center gap-x-2 gap-y-1"
            >
              {#each missingRequiredFilters as missing (missing.key)}
                <li
                  class="px-2 py-0.5 rounded-md bg-red-50 border border-red-200 text-red-700"
                >
                  {missing.label}
                </li>
              {/each}
            </ul>
          </div>
        </div>
      {:else}
        <div
          class="w-full h-fit flex flex-col items-center row-container relative"
          style:max-width="{maxWidth}px"
          style:min-width="420px"
          bind:contentRect
        >
          <slot />
        </div>
      {/if}
    </div>

    <CellInspector />
  </main>
</ThemeProvider>

<style>
  div {
    container-type: inline-size;
    container-name: canvas-container;
  }
</style>
