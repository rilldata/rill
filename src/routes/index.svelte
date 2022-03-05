<script>
import Workspace from "./_surfaces/workspace/index.svelte";
import InspectorSidebar from "./_surfaces/inspector/index.svelte";
import AssetsSidebar from "./_surfaces/assets/index.svelte";
import Header from "./_surfaces/header/index.svelte";
import { setContext } from "svelte";

import PaneExpanderIcon from "$lib/components/PaneExpanderIcon.svelte";

import SurfaceCollapseButton from "$lib/components/SurfaceCollapseButton.svelte"

import { 
  layout,
  assetVisibilityTween, 
  assetsVisible,
  inspectorVisibilityTween,
  inspectorVisible
} from "$lib/layout-store"
import Portal from "$lib/components/Portal.svelte";

setContext("rill:app:layout", layout);

let leftHovered = false;
let rightHovered = false;
let elementHovered = false;

</script>

<div class='body'>

  <!-- left expansion button -->
  <!-- make this the first element to select with tab by placing it first.-->
  <SurfaceCollapseButton
    show={(leftHovered || !$assetsVisible)}
    left="{($layout.assetsWidth - 12 - 24) * (1 - $assetVisibilityTween) + 12 * $assetVisibilityTween}px"
    on:click={() => {
      assetsVisible.set($assetsVisible ? 0 : 1);
    }}
  >
    <PaneExpanderIcon size="16px" mode={$assetsVisible ? "right" : 'hamburger'} />
    <svelte:fragment slot="tooltip-content">
      {#if $assetVisibilityTween === 0} hide {:else} show {/if} models and tables
    </svelte:fragment>
  </SurfaceCollapseButton>

  <div class="surface assets fixed"
    aria-hidden={!$assetsVisible}
    on:mouseover={() => { leftHovered = true; }}
    on:mouseleave={() => { leftHovered = false; }}
    on:focus={() => { leftHovered = true; }}
    on:blur={() => { leftHovered = false; }}
    style:left="{-$assetVisibilityTween * $layout.assetsWidth}px"
  >

    <AssetsSidebar />
  </div>  
  
  <div 
    class="surface inputs bg-gray-100 fixed" 
    style:padding-left="{($assetVisibilityTween * 80)}px"
    style:padding-right="{($inspectorVisibilityTween * 80)}px"
    style:left="{$layout.assetsWidth * (1 - $assetVisibilityTween)}px" 
    style:top="0px" 
    style:right="{$layout.inspectorWidth * (1 - $inspectorVisibilityTween)}px">
    <Header />
    <Workspace />
  </div>

  <!-- inspector  collapse button should be tabbable as if it were the first element. -->
  <SurfaceCollapseButton
    show={rightHovered || !$inspectorVisible}
    right="{($layout.inspectorWidth - 12 - 24) * (1 - $inspectorVisibilityTween) + 12 * $inspectorVisibilityTween}px"
    on:click={() => {
      inspectorVisible.set($inspectorVisible ? 0 : 1);
    }}
  >
    <PaneExpanderIcon size="16px" mode={$inspectorVisible ? "left" : 'right'} />
    <svelte:fragment slot="tooltip-content">
      {#if $inspectorVisibilityTween === 0} hide {:else} show {/if} the model inspector
    </svelte:fragment>
  </SurfaceCollapseButton>

  <div 
    class='fixed'
    aria-hidden={!$inspectorVisible}
    on:mouseover={() => { rightHovered = true; }}
    on:mouseleave={() => { rightHovered = false; }}
    on:focus={() => { rightHovered = true; }}
    on:blur={() => { rightHovered = false; }}
    style:right="{$layout.inspectorWidth * (1 - $inspectorVisibilityTween)}px" 
  >
    <InspectorSidebar />
  </div>

</div>
<style>

.body {
    width: 100vw;
    position:absolute;
    height: calc(100vh);
  }
.inputs {
  --hue: 217;
  --sat: 20%;
  --lgt: 95%;
  --bg: hsl(var(--hue), var(--sat), var(--lgt));
  --bg-transparent: hsla(var(--hue), var(--sat), var(--lgt), .8);
  /* background-color: var(--bg); */
  overflow-y: auto;
  height:100%;
}

.surface {
  box-sizing: border-box;
}

.surface:first-child {
  border-right: 1px solid #ddd;
}

.outputs {
  overflow-y: auto;
  height:100%;
}

.surface.outputs, .surface.assets {
  overflow-y: auto;
  overflow-x: hidden;
}

.preview-drawer {
  overflow: hidden;
}

</style>