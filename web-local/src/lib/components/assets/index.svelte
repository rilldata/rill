<script lang="ts">
  import { onMount } from "svelte";
  import {
    assetsVisible,
    assetVisibilityTween,
    layout,
  } from "../../application-state-stores/layout-store";
  import { drag } from "../../drag";
  import RillLogo from "../icons/RillLogo.svelte";
  import Spacer from "../icons/Spacer.svelte";
  import Portal from "../Portal.svelte";
  import Footer from "./Footer.svelte";

  import MetricsDefinitionAssets from "./MetricsDefinitionAssets.svelte";
  import ModelAssets from "./ModelAssets.svelte";
  import TableAssets from "./TableAssets.svelte";

  let mounted = false;
  onMount(() => {
    mounted = true;
  });
</script>

<div
  class="
  border-r 
  border-transparent 
  fixed 
  overflow-auto 
  border-gray-200 
  transition-colors
  h-screen
  bg-white
"
  class:hidden={$assetVisibilityTween === 1}
  class:pointer-events-none={!$assetsVisible}
  style:top="0px"
  style:width="{$layout.assetsWidth}px"
>
  <!-- draw handler -->
  {#if $assetsVisible}
    <Portal>
      <div
        class="fixed z-50 drawer-handler w-4 hover:cursor-col-resize -translate-x-2 h-screen"
        style:left="{(1 - $assetVisibilityTween) * $layout.assetsWidth}px"
        use:drag={{ minSize: 300, maxSize: 500, side: "assetsWidth" }}
      />
    </Portal>
  {/if}

  <div class="w-full flex flex-col h-full">
    <div class="grow" style:outline="1px solid black">
      <header
        style:height="var(--header-height)"
        class="sticky top-0 grid align-center bg-white z-50"
      >
        <h1
          class="grid grid-flow-col justify-start gap-x-3 p-4 items-center content-center"
        >
          {#if mounted}
            <a href="/">
              <RillLogo size="16px" iconOnly />
            </a>
          {:else}
            <Spacer size="16px" />
          {/if}
          <a href="/" class="font-bold text-black"> Rill Developer </a>
        </h1>
      </header>
      <TableAssets />
      <ModelAssets />
      <MetricsDefinitionAssets />
    </div>
    <Footer />
  </div>
</div>
