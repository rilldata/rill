<script lang="ts">
  import HideLeftSidebar from "@rilldata/web-common/components/icons/HideLeftSidebar.svelte";
  import SurfaceViewIcon from "@rilldata/web-common/components/icons/SurfaceView.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { ModelAssets } from "@rilldata/web-common/features/models";
  import ProjectTitle from "@rilldata/web-common/features/project/ProjectTitle.svelte";
  import SourceAssets from "@rilldata/web-common/features/sources/navigation/SourceAssets.svelte";
  import { getContext } from "svelte";
  import { tweened } from "svelte/motion";
  import { Readable, Writable, writable } from "svelte/store";
  import ChartAssets from "../../features/charts/ChartAssets.svelte";
  import CustomDashboardAssets from "../../features/custom-dashboards/CustomDashboardAssets.svelte";
  import DashboardAssets from "../../features/dashboards/DashboardAssets.svelte";
  import OtherFiles from "../../features/project/OtherFiles.svelte";
  import TableAssets from "../../features/tables/TableAssets.svelte";
  import { useIsModelingSupportedForCurrentOlapDriver } from "../../features/tables/selectors";
  import { createRuntimeServiceGetInstance } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { DEFAULT_NAV_WIDTH } from "../config";
  import { drag } from "../drag";
  import Footer from "./Footer.svelte";
  import SurfaceControlButton from "./SurfaceControlButton.svelte";
  import { portal } from "@rilldata/web-common/lib/actions/portal";

  const { customDashboards } = featureFlags;

  /** FIXME: come up with strong defaults here when needed */
  const navigationLayout =
    getContext<
      Writable<{
        value: number;
        visible: boolean;
      }>
    >("rill:app:navigation-layout") ||
    writable({ value: DEFAULT_NAV_WIDTH, visible: true });

  const navigationWidth =
    getContext<Readable<number>>("rill:app:navigation-width-tween") ||
    writable(DEFAULT_NAV_WIDTH);

  const navVisibilityTween =
    getContext<Readable<number>>("rill:app:navigation-visibility-tween") ||
    tweened(0, { duration: 50 });

  const { readOnly } = featureFlags;

  let previousWidth: number;

  $: isModelerEnabled = $readOnly === false;

  $: instance = createRuntimeServiceGetInstance($runtime.instanceId);
  $: olapConnector = $instance.data?.instance?.olapConnector;
  $: isModelingSupportedForCurrentOlapDriver =
    useIsModelingSupportedForCurrentOlapDriver($runtime.instanceId);

  function handleResize(
    e: UIEvent & {
      currentTarget: EventTarget & Window;
    },
  ) {
    const currentWidth = e.currentTarget.innerWidth;

    if (currentWidth < previousWidth && currentWidth < 768) {
      $navigationLayout.visible = false;
    }

    previousWidth = currentWidth;
  }
</script>

<svelte:window on:resize={handleResize} />

<div
  aria-hidden={!$navigationLayout?.visible}
  class="box-border assets fixed"
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
      <div
        use:portal
        role="separator"
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
          store: navigationLayout,
        }}
      />
    {/if}

    <div class="w-full flex flex-col h-full">
      <div class="grow">
        <ProjectTitle />
        {#if isModelerEnabled}
          <TableAssets />
          {#if olapConnector === "duckdb"}
            <SourceAssets />
          {/if}
          {#if $isModelingSupportedForCurrentOlapDriver.data}
            <ModelAssets />
          {/if}
        {/if}
        <DashboardAssets />
        {#if $customDashboards}
          <ChartAssets />
          <CustomDashboardAssets />
        {/if}
        {#if isModelerEnabled}
          <OtherFiles />
        {/if}
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
    {#if $navVisibilityTween === 0}
      Close
    {:else}
      Show
    {/if} sidebar
  </svelte:fragment>
</SurfaceControlButton>

<style>
</style>
