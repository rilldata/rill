<script>
import Workspace from "./_surfaces/workspace/index.svelte";
import InspectorSidebar from "./_surfaces/inspector/index.svelte";
import AssetsSidebar from "./_surfaces/assets/index.svelte";
import Header from "./_surfaces/workspace/Header.svelte";

import SurfaceViewIcon from "$lib/components/icons/SurfaceView.svelte";

import SurfaceControlButton from "$lib/components/surface/SurfaceControlButton.svelte"

import { 
  layout,
  assetVisibilityTween, 
  assetsVisible,
  inspectorVisibilityTween,
  inspectorVisible,
  SIDE_PAD
} from "$lib/layout-store"


let assetsHovered = false;
let inspectorHovered = false;

</script>

<div class="absolute w-screen h-screen">

  <!-- left assets pane expansion button -->
  <!-- make this the first element to select with tab by placing it first.-->
  <SurfaceControlButton
    show={(assetsHovered || !$assetsVisible)}
    left="{($layout.assetsWidth - 12 - 24) * (1 - $assetVisibilityTween) + 12 * $assetVisibilityTween}px"
    on:click={() => {
      assetsVisible.set($assetsVisible ? 0 : 1);
    }}
  >
    <SurfaceViewIcon size="16px" mode={$assetsVisible ? "right" : 'hamburger'} />
    <svelte:fragment slot="tooltip-content">
      {#if $assetVisibilityTween === 0} hide {:else} show {/if} models and tables
    </svelte:fragment>
  </SurfaceControlButton>

  <!-- assets sidebar component -->
  <!-- this is where we handle navigation -->
  <div class="box-border	 assets fixed"
    aria-hidden={!$assetsVisible}
    on:mouseover={() => { assetsHovered = true; }}
    on:mouseleave={() => { assetsHovered = false; }}
    on:focus={() => { assetsHovered = true; }}
    on:blur={() => { assetsHovered = false; }}
    style:left="{-$assetVisibilityTween * $layout.assetsWidth}px"
  >
    <AssetsSidebar />
  </div>  
  
  <!-- workspace component -->
  <div 
    class="box-border bg-gray-100 fixed" 
    style:padding-left="{($assetVisibilityTween * SIDE_PAD)}px"
    style:padding-right="{($inspectorVisibilityTween * SIDE_PAD)}px"
    style:left="{$layout.assetsWidth * (1 - $assetVisibilityTween)}px" 
    style:top="0px" 
    style:right="{$layout.inspectorWidth * (1 - $inspectorVisibilityTween)}px">
    <Header />
    <Workspace />
  </div>

  <!-- inspector collapse button should be tabbable as if it were the first element of the inspector. -->
  <SurfaceControlButton
    show={inspectorHovered || !$inspectorVisible}
    right="{($layout.inspectorWidth - 12 - 24) * (1 - $inspectorVisibilityTween) + 12 * $inspectorVisibilityTween}px"
    on:click={() => {
      inspectorVisible.set($inspectorVisible ? 0 : 1);
    }}
  >
    <SurfaceViewIcon size="16px" mode={$inspectorVisible ? "left" : 'right'} />
    <svelte:fragment slot="tooltip-content">
      {#if $inspectorVisibilityTween === 0} hide {:else} show {/if} the model inspector
    </svelte:fragment>
  </SurfaceControlButton>

  <!-- inspector sidebar -->
  <div 
    class='fixed'
    aria-hidden={!$inspectorVisible}
    on:mouseover={() => { inspectorHovered = true; }}
    on:mouseleave={() => { inspectorHovered = false; }}
    on:focus={() => { inspectorHovered = true; }}
    on:blur={() => { inspectorHovered = false; }}
    style:right="{$layout.inspectorWidth * (1 - $inspectorVisibilityTween)}px" 
  >
    <InspectorSidebar />
  </div>
</div>