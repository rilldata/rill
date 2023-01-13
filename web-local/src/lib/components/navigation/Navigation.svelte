<script lang="ts">
  import HideLeftSidebar from "@rilldata/web-common/components/icons/HideLeftSidebar.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import SurfaceViewIcon from "@rilldata/web-common/components/icons/SurfaceView.svelte";
  import Portal from "@rilldata/web-common/components/Portal.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { ModelAssets } from "@rilldata/web-common/features/models";
  import TableAssets from "@rilldata/web-common/features/sources/navigation/TableAssets.svelte";
  import { useRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import SurfaceControlButton from "@rilldata/web-local/lib/components/surface/SurfaceControlButton.svelte";
  import { getContext, onMount } from "svelte";
  import { tweened } from "svelte/motion";
  import { Readable, Writable, writable } from "svelte/store";
  import { parseDocument } from "yaml";
  import { DEFAULT_NAV_WIDTH } from "../../application-config";
  import { drag } from "../../drag";
  import MetricsDefinitionAssets from "./dashboards/MetricsDefinitionAssets.svelte";
  import Footer from "./Footer.svelte";
  import { shorthandTitle } from "./shorthand-title";

  let mounted = false;
  onMount(() => {
    mounted = true;
  });

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

  $: thing = useRuntimeServiceGetFile(
    $runtimeStore?.instanceId,
    `rill.yaml`
    //getFilePathFromNameAndType(metricsDefName, EntityType.MetricsDefinition)
  );

  $: yaml = parseDocument($thing?.data?.blob || "{}")?.toJS();
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
        <header
          style:height="var(--header-height)"
          class="sticky top-0 grid align-center bg-white z-50"
        >
          <!-- the pl-[.875rem] is a fix to move this new element over a pinch.-->
          <h1
            class="grid grid-flow-col justify-start gap-x-3 p-4 pl-[.875rem] items-center content-center"
          >
            {#if mounted}
              <a href="/">
                <div
                  style:width="20px"
                  style:font-size="9px"
                  class="grid place-items-center rounded bg-gray-800 text-white font-normal"
                  style:height="20px"
                >
                  <div>
                    {shorthandTitle(yaml?.name || "Ri")}
                  </div>
                </div>
              </a>
            {:else}
              <Spacer size="16px" />
            {/if}
            <Tooltip distance={8}>
              <a
                href="/"
                class="font-semibold text-black grow text-ellipsis overflow-hidden whitespace-nowrap pr-12"
              >
                {yaml?.name || "Untitled Rill Project"}
              </a>
              <TooltipContent maxWidth="300px" slot="tooltip-content">
                <div class="font-bold">
                  {yaml?.name || "Untitled Rill Project"}
                </div>
              </TooltipContent>
            </Tooltip>
          </h1>
        </header>
        <MetricsDefinitionAssets />
        <TableAssets />
        <ModelAssets />
      </div>
      <Footer />
    </div>
  </div>
</div>

<SurfaceControlButton
  left="{($navigationWidth - 12 - 24) * (1 - $navVisibilityTween) +
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
