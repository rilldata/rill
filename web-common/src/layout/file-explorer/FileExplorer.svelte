<script lang="ts">
  import ProjectTitle from "@rilldata/web-common/features/project/ProjectTitle.svelte";
  import { navigationOpen } from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import { DEFAULT_NAV_WIDTH } from "../config";
  import Footer from "../navigation/Footer.svelte";
  import Resizer from "../Resizer.svelte";
  import SurfaceControlButton from "../navigation/SurfaceControlButton.svelte";
  import Folder from "./Folder.svelte";
  import type { DirectoryOrFile } from "./createMap";

  export let files: DirectoryOrFile;

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
    <div class="grow overflow-auto">
      <Folder name="Rill Project" {files} expanded />
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

  .sidebar:not(.resizing) {
    transition-duration: 400ms;
    transition-timing-function: ease-in-out;
  }

  .hide {
    width: 0px !important;
  }
</style>
