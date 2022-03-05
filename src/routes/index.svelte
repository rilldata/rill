<script>
import Workspace from "./_surfaces/workspace/index.svelte";
import InspectorSidebar from "./_surfaces/inspector/index.svelte";
import AssetsSidebar from "./_surfaces/assets/index.svelte";
import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
import Header from "./_surfaces/header/index.svelte";
import { setContext } from "svelte";
import { fly, fade } from "svelte/transition";
import PaneExpanderIcon from "$lib/components/PaneExpanderIcon.svelte";
import { panes } from "$lib/pane-store"
import { 
  assetVisibilityTween, 
  assetsVisible,
  inspectorVisibilityTween,
  inspectorVisible
} from "$lib/pane-store"

setContext("rill:app:panes", panes);

let leftHovered = false;
let rightHovered = false;
let elementHovered = false;

</script>

<div class='body'>
  <div class="surface assets fixed"
    on:mouseover={() => { leftHovered = true; }}
    on:mouseleave={() => { leftHovered = false; }}
    on:focus={() => { leftHovered = true; }}
    on:blur={() => { leftHovered = false; }}
    style:left="{-$assetVisibilityTween * $panes.left}px"
  >
    <AssetsSidebar />
  </div>  
  
  <div 
    class="surface inputs bg-gray-100 fixed" 
    style:padding-left="{($assetVisibilityTween * 80)}px"
    style:padding-right="{($inspectorVisibilityTween * 80)}px"
    style:left="{$panes.left * (1 - $assetVisibilityTween)}px" 
    style:top="0px" 
    style:right="{$panes.right * (1 - $inspectorVisibilityTween)}px">
    <Header />
    <Workspace />
  </div>

  <div 
    class='
      
      fixed
    '
    on:mouseover={() => { rightHovered = true; }}
    on:mouseleave={() => { rightHovered = false; }}
    on:focus={() => { rightHovered = true; }}
    on:blur={() => { rightHovered = false; }}
    style:right="{$panes.right * (1- $inspectorVisibilityTween)}px" 
  >
    <InspectorSidebar />
  </div>

<div>
    <Tooltip location="right" alignment="center" distance={12}>
      <button 
        class="fixed z-40  {leftHovered || !$assetsVisible ? "opacity-100" : "opacity-0"} hover:opacity-100 transition-opacity"
        style:left="{($panes.left - 12 - 24) * (1 - $assetVisibilityTween) + 12 * $assetVisibilityTween}px"
        style:top="12px"
        on:click={() => {
          assetsVisible.set($assetsVisible ? 0 : 1);
      }}>
      <div 
      class="rounded bg-transparent hover:bg-gray-300 transition-colors grid place-items-center text-gray-500 hover:text-gray-800"
        style:width="24px" 
        style:height="24px" 
      >
        <PaneExpanderIcon size="16px" mode={$assetsVisible ? "right" : 'hamburger'} />
      </div>
    </button>
  <TooltipContent slot="tooltip-content">
    {#if $assetVisibilityTween === 0} hide {:else} show {/if} models and tables
  </TooltipContent>
  </Tooltip>



  <Tooltip location="left" alignment="center" distance={12}>
    <button 
    class="fixed z-40  {rightHovered || !$inspectorVisible ? "opacity-100" : "opacity-0"} hover:opacity-100 transition-opacity"
    style:right="{($panes.right - 12 - 24) * (1 - $inspectorVisibilityTween) + 12 * $inspectorVisibilityTween}px"
      style:top="12px"
      on:click={() => {
        inspectorVisible.set($inspectorVisible ? 0 : 1);
    }}>
    <div 
      class="rounded bg-transparent hover:bg-gray-300 transition-colors grid place-items-center text-gray-500 hover:text-gray-800"
      style:width="24px" 
      style:height="24px" 

      >
      <PaneExpanderIcon size="16px" mode={$inspectorVisible ? "left" : 'right'} />
    </div>
  </button>
<TooltipContent slot="tooltip-content">
  {#if $inspectorVisibilityTween === 0} hide {:else} show {/if} the model inspector
</TooltipContent>
</Tooltip>

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