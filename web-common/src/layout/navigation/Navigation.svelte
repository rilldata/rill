<script lang="ts">
  import HideLeftSidebar from "@rilldata/web-common/components/icons/HideLeftSidebar.svelte";
  import SurfaceViewIcon from "@rilldata/web-common/components/icons/SurfaceView.svelte";
  import Portal from "@rilldata/web-common/components/Portal.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { ModelAssets } from "@rilldata/web-common/features/models";
  import TableAssets from "@rilldata/web-common/features/sources/navigation/TableAssets.svelte";
  import ProjectTitle from "@rilldata/web-common/layout/navigation/ProjectTitle.svelte";
  import { getContext } from "svelte";
  import { tweened } from "svelte/motion";
  import { Readable, Writable, writable } from "svelte/store";
  import DashboardAssets from "../../features/dashboards/DashboardAssets.svelte";
  import { DEFAULT_NAV_WIDTH } from "../config";
  import { drag } from "../drag";
  import Footer from "./Footer.svelte";
  import SurfaceControlButton from "./SurfaceControlButton.svelte";

  /** FIXME: come up with strong defaults here when needed */
  const navigationLayout =
    (getContext("rill:app:navigation-layout") as Writable<{
      value: number;
      visible: boolean;
    }>) || writable({ value: DEFAULT_NAV_WIDTH, visible: true });

  const navigationWidth =
    (getContext("rill:app:navigation-width-tween") as Readable<number>) ||
    writable(DEFAULT_NAV_WIDTH);

  const navVisibilityTween =
    (getContext("rill:app:navigation-visibility-tween") as Readable<number>) ||
    tweened(0, { duration: 50 });

  $: isModelerEnabled = $featureFlags.readOnly === false;
</script>

<div
  aria-hidden={!$navigationLayout?.visible}
  class="box-border	assets fixed"
  style:left="{-$navVisibilityTween * $navigationWidth}px"
>
  <div
    class="
  border-r 
  fixed 
  overflow-auto 
  border-gray-200 
  transition-colors
  h-screen
  bg-white
"
    class:hidden={$navVisibilityTween === 1}
    class:pointer-events-none={!$navigationLayout?.visible}
    style:top="0px"
    style:width="{$navigationWidth}px"
  >
    <!-- draw handler -->
    {#if $navigationLayout?.visible}
      <Portal>
        <div
          on:dblclick={() => {
            navigationLayout.update((state) => {
              state.value = DEFAULT_NAV_WIDTH;
              return state;
            });
          }}
          class="fixed drawer-handler w-4 hover:cursor-col-resize -translate-x-2 h-screen"
          style:left="{(1 - $navVisibilityTween) * $navigationWidth}px"
          use:drag={{
            minSize: DEFAULT_NAV_WIDTH,
            maxSize: 440,
            side: "assetsWidth",
            store: navigationLayout,
          }}
        />
      </Portal>
    {/if}

    <div class="w-full flex flex-col h-full">
      <div class="grow">
        <ProjectTitle />
        {#if isModelerEnabled}
          <TableAssets />
          <ModelAssets />
        {/if}
        <DashboardAssets />
      </div>
      <Footer />
    </div>
  </div>
</div>

<SurfaceControlButton
  left="{($navigationWidth - 12 - 20) * (1 - $navVisibilityTween) +
    12 * $navVisibilityTween}px"
  on:click={() => {
    //assetsVisible.set(!$assetsVisible);
    navigationLayout.update((state) => {
      state.visible = !state.visible;
      return state;
    });
  }}
  show={true}
>
  {#if $navigationLayout?.visible}
    <HideLeftSidebar size="18px" />
  {:else}
    <SurfaceViewIcon size="16px" mode={"hamburger"} />
  {/if}
  <svelte:fragment slot="tooltip-content">
    {#if $navVisibilityTween === 0} Close {:else} Show {/if} sidebar
  </svelte:fragment>
</SurfaceControlButton>
