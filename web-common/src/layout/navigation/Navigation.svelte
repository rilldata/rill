<script lang="ts" context="module">
  export const navigationOpen = (() => {
    const { subscribe, update, set } = writable<boolean | null>(true);
    return {
      toggle: () => update((open) => !open),
      set,
      subscribe,
    };
  })();
</script>

<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { connectorExplorerStore } from "@rilldata/web-common/features/connectors/explorer/connector-explorer-store";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { writable } from "svelte/store";
  import ConnectorExplorer from "../../features/connectors/explorer/ConnectorExplorer.svelte";
  import AddAssetButton from "../../features/entity-management/AddAssetButton.svelte";
  import FileExplorer from "../../features/file-explorer/FileExplorer.svelte";
  import Resizer from "../Resizer.svelte";
  import { DEFAULT_NAV_WIDTH, MAX_NAV_WIDTH, MIN_NAV_WIDTH } from "../config";
  import Footer from "./Footer.svelte";
  import SurfaceControlButton from "./SurfaceControlButton.svelte";

  const DEFAULT_PERCENTAGE = 0.4;

  let width = DEFAULT_NAV_WIDTH;
  let previousWidth: number;
  let resizing = false;
  let resizingConnector = false;
  let connectorHeightPercentage = DEFAULT_PERCENTAGE;
  let contentRect = new DOMRectReadOnly(0, 0, 0, 0);
  let connectorWrapper: HTMLDivElement;

  $: navWrapperHeight = contentRect.height;

  let showConnectors = true;

  $: connectorSectionHeight = navWrapperHeight * connectorHeightPercentage;

  $: ({ unsavedFiles } = fileArtifacts);
  $: ({ size: unsavedFileCount } = $unsavedFiles);

  function handleResize(
    e: UIEvent & {
      currentTarget: EventTarget & Window;
    },
  ) {
    const currentWidth = e.currentTarget.innerWidth;

    const open = $navigationOpen;

    if (open && currentWidth < previousWidth && currentWidth < 768) {
      $navigationOpen = null;
    } else if (open === null && currentWidth > 768) {
      $navigationOpen = true;
    }

    previousWidth = currentWidth;
  }
</script>

<svelte:window
  on:resize={handleResize}
  on:keydown={(e) => {
    const isMac = window.navigator.userAgent.includes("Macintosh");

    if (e[isMac ? "metaKey" : "ctrlkey"] && e.key === "b") {
      navigationOpen.toggle();
    }
  }}
/>

<nav
  class="sidebar"
  class:hide={!$navigationOpen}
  class:resizing
  style:width="{width}px"
>
  <Resizer
    min={MIN_NAV_WIDTH}
    basis={DEFAULT_NAV_WIDTH}
    max={MAX_NAV_WIDTH}
    dimension={width}
    onUpdate={(w) => {
      width = w;
    }}
    bind:resizing
    side="right"
  />
  <div class="inner" style:width="{width}px">
    <div class="p-2 w-full pr-10">
      <AddAssetButton />
    </div>
    <div class="scroll-container">
      <div class="nav-wrapper" bind:contentRect>
        <section class="size-full overflow-y-auto pb-4">
          <FileExplorer hasUnsaved={unsavedFileCount > 0} />
        </section>

        {#if navWrapperHeight}
          <section class="connector-section">
            {#if showConnectors}
              <Resizer
                dimension={connectorSectionHeight}
                onUpdate={(height) => {
                  connectorHeightPercentage = height / navWrapperHeight;
                }}
                direction="NS"
                side="top"
                min={0}
                basis={navWrapperHeight * DEFAULT_PERCENTAGE}
                max={navWrapperHeight * 0.9}
                bind:resizing={resizingConnector}
              />
            {/if}

            <button
              on:click={() => {
                const open = showConnectors;

                if (!open) showConnectors = true;

                connectorWrapper.animate(
                  [
                    {
                      height: `${open ? connectorSectionHeight : 0}px`,
                    },
                    {
                      height: `${open ? 0 : connectorSectionHeight}px`,
                    },
                  ],
                  {
                    duration: 200,
                    easing: "ease-out",
                  },
                ).onfinish = () => {
                  if (open) showConnectors = false;
                };
              }}
            >
              <CaretDownIcon
                size="14px"
                className="text-gray-400 transition-transform {!showConnectors &&
                  '-rotate-90'}"
              />
              <h3>Data Explorer</h3>
            </button>

            <div
              class="connector-wrapper"
              role="region"
              aria-label="Data explorer"
              bind:this={connectorWrapper}
              style:height="{showConnectors ? connectorSectionHeight : 0}px"
            >
              {#if showConnectors}
                <ConnectorExplorer store={connectorExplorerStore} />
              {/if}
            </div>
          </section>
        {/if}
      </div>
    </div>
    <Footer />
  </div>
</nav>

<SurfaceControlButton
  {resizing}
  navWidth={width}
  navOpen={!!$navigationOpen}
  onClick={navigationOpen.toggle}
/>

<style lang="postcss">
  .sidebar {
    @apply flex flex-col flex-none relative overflow-hidden bg-surface;
    @apply h-full border-r z-0;
    @apply select-none;
    transition-property: width;
    will-change: width;
  }

  .inner {
    @apply h-full overflow-hidden flex flex-col;
    will-change: width;
  }

  .nav-wrapper {
    @apply flex flex-col size-full;
  }

  .scroll-container {
    @apply overflow-y-auto overflow-x-hidden;
    @apply h-full bg-surface;
  }

  .sidebar:not(.resizing) {
    transition-duration: 300ms;
    transition-timing-function: ease-in-out;
  }

  .hide {
    width: 0px !important;
  }

  .connector-section {
    @apply flex flex-col flex-none h-fit;
    @apply border-t relative;
  }

  .connector-wrapper {
    @apply overflow-y-auto;
  }

  button {
    @apply flex gap-x-1 items-center w-full;
    @apply pl-2 pr-3.5 py-1.5 cursor-pointer;
    @apply text-gray-500;
  }

  button:hover {
    @apply bg-slate-100;
  }

  h3 {
    @apply font-semibold text-[10px] uppercase;
  }
</style>
