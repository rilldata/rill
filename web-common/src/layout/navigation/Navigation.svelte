<script lang="ts" context="module">
  export const navigationOpen = (() => {
    const store = writable(true);
    return {
      subscribe: store.subscribe,
      toggle: () => store.update((open) => !open),
    };
  })();
</script>

<script lang="ts">
  import ProjectTitle from "@rilldata/web-common/features/project/ProjectTitle.svelte";
  import { writable } from "svelte/store";
  import AddAssetButton from "../../features/entity-management/AddAssetButton.svelte";
  import FileExplorer from "../../features/file-explorer/FileExplorer.svelte";
  import TableAssets from "../../features/tables/TableAssets.svelte";
  import Resizer from "../Resizer.svelte";
  import { DEFAULT_NAV_WIDTH } from "../config";
  import Footer from "./Footer.svelte";
  import SurfaceControlButton from "./SurfaceControlButton.svelte";

  let width = DEFAULT_NAV_WIDTH;
  let previousWidth: number;
  let container: HTMLElement;
  let resizing = false;

  function handleResize(
    e: UIEvent & {
      currentTarget: EventTarget & Window;
    },
  ) {
    const currentWidth = e.currentTarget.innerWidth;

    if (currentWidth < previousWidth && currentWidth < 768) {
      $navigationOpen = false;
    }

    previousWidth = currentWidth;
  }
</script>

<svelte:window on:resize={handleResize} />

<nav
  class="sidebar"
  class:hide={!$navigationOpen}
  class:resizing
  style:width="{width}px"
  bind:this={container}
>
  <Resizer
    min={DEFAULT_NAV_WIDTH}
    basis={DEFAULT_NAV_WIDTH}
    max={440}
    bind:dimension={width}
    bind:resizing
    side="right"
  />
  <div class="inner" style:width="{width}px">
    <ProjectTitle />

    <AddAssetButton />
    <div class="scroll-container">
      <div class="nav-wrapper">
        <FileExplorer />
        <TableAssets />
      </div>
    </div>
    <Footer />
  </div>
</nav>

<SurfaceControlButton
  {resizing}
  navWidth={width}
  navOpen={$navigationOpen}
  on:click={navigationOpen.toggle}
/>

<style lang="postcss">
  .sidebar {
    @apply flex flex-col flex-none relative overflow-hidden;
    @apply h-screen border-r z-0;
    transition-property: width;
    will-change: width;
  }

  .inner {
    @apply h-full overflow-hidden flex flex-col;
    will-change: width;
  }

  .nav-wrapper {
    @apply flex flex-col h-fit w-full gap-y-2;
  }

  .scroll-container {
    @apply overflow-y-auto overflow-x-hidden;
    @apply transition-colors h-full bg-white pb-8;
  }

  .sidebar:not(.resizing) {
    transition-duration: 400ms;
    transition-timing-function: ease-in-out;
  }

  .hide {
    width: 0px !important;
  }
</style>
