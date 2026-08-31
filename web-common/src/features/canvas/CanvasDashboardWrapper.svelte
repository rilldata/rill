<script lang="ts">
  import { dynamicHeight } from "@rilldata/web-common/layout/layout-settings.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import CellInspector from "@rilldata/web-common/components/CellInspector.svelte";
  import CanvasFilters from "./filters/CanvasFilters.svelte";
  import { getCanvasStore } from "./state-managers/state-managers";
  import ThemeProvider from "../dashboards/ThemeProvider.svelte";
  import CanvasPdfExportView from "../exports/pdf/CanvasPdfExportView.svelte";
  import { getMissingRequiredFilters } from "@rilldata/web-common/features/dashboards/filters/utils.ts";
  import MissingRequiredFiltersMessage from "@rilldata/web-common/features/dashboards/filters/MissingRequiredFiltersMessage.svelte";
  import { type Snippet } from "svelte";
  import { syncStoreWithSource } from "@rilldata/web-common/lib/store-utils/url-params-store-sync.svelte.ts";
  import { goto } from "$app/navigation";

  const runtimeClient = useRuntimeClient();
  let instanceId = $derived(runtimeClient.instanceId);

  let {
    children,
    canvasName,
    maxWidth,
    clientWidth = $bindable(0),
    showGrabCursor = false,
    filtersEnabled = undefined,
    embedded = false,
    builder = false,
    onClick = () => {},
  }: {
    children: Snippet;
    canvasName: string;
    maxWidth: number;
    clientWidth?: number;
    showGrabCursor?: boolean;
    filtersEnabled?: boolean | undefined;
    embedded?: boolean;
    builder?: boolean;
    onClick?: () => void;
  } = $props();

  let contentRect = $state(new DOMRectReadOnly(0, 0, 0, 0));

  let {
    canvasEntity: {
      theme,
      exportMode,
      expressionFilterManager,
      dashboardProvider,
    },
  } = $derived(getCanvasStore(canvasName, instanceId));
  // svelte-ignore state_referenced_locally
  expressionFilterManager.createListener();
  // svelte-ignore state_referenced_locally
  syncStoreWithSource(
    expressionFilterManager,
    (newUrlParams) => goto("?" + newUrlParams.toString()),
    () => expressionFilterManager.metricsViewsProvider.ready,
  );

  $effect(() => {
    dashboardProvider.yamlConfigProvider.setEditable(builder);
  });

  let missingRequiredFilters = $derived(
    getMissingRequiredFilters(
      expressionFilterManager,
      dashboardProvider.metricsViewsProvider,
      dashboardProvider.yamlConfigProvider,
    ),
  );
  // Note that this can be false when metricsViewsProvider is resolving.
  // TODO: make sure parent waits for metricsViewsProvider.ready
  let hasMissingRequired = $derived(missingRequiredFilters.length > 0);

  $effect(() => {
    clientWidth = contentRect.width;
  });
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
        <CanvasFilters {canvasName} {maxWidth} />
      </header>
    {/if}

    <!-- Off-screen, read-only full render of the canvas, mounted solely during a
         PDF export as the capture target (see CanvasPdfExportView). Kept out of
         the DOM otherwise: Playwright locators and the a11y tree match elements
         regardless of CSS visibility, so an always-mounted copy would duplicate
         the live dashboard's text/labels. -->
    {#if $exportMode && !hasMissingRequired}
      <div
        aria-hidden="true"
        class="pointer-events-none absolute"
        style="left: -99999px; top: 0;"
      >
        <CanvasPdfExportView
          {canvasName}
          {instanceId}
          width={clientWidth || maxWidth}
        />
      </div>
    {/if}

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
        <MissingRequiredFiltersMessage
          {missingRequiredFilters}
          metricsViewsProvider={dashboardProvider.metricsViewsProvider}
        />
      {:else}
        <div
          class="w-full h-fit flex flex-col items-center row-container relative"
          style:max-width="{maxWidth}px"
          style:min-width="420px"
          bind:contentRect
        >
          {@render children()}
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
