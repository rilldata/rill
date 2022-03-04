<script>
import Workspace from "./_surfaces/workspace/index.svelte";
import InspectorSidebar from "./_surfaces/inspector/index.svelte";
import AssetsSidebar from "./_surfaces/assets/index.svelte";
import PreviewDrawer from "./_surfaces/preview/index.svelte";
import Header from "./_surfaces/header/index.svelte";
import { setContext } from "svelte";
import { fly } from "svelte/transition";
import { panes } from "$lib/pane-store"
import { 
  assetVisibilityTween, 
  assetsVisible,
  inspectorVisibilityTween,
  inspectorVisible
} from "$lib/pane-store"

setContext("rill:app:panes", panes);

</script>

<div class='body'>
  <div class="surface assets" style="
    position: fixed;
  "
  style:left="{-$assetVisibilityTween * $panes.left}px"
  >
    <AssetsSidebar />
  </div>  
  
  <div 
    class="surface inputs bg-gray-100 fixed" 
    style:padding-left="{($assetVisibilityTween * 80)}px"
    style:left="{$panes.left * (1 - $assetVisibilityTween)}px" 
    style:top="0px" 
    style:right="{$panes.right}px">
    {#if !$assetsVisible}
      <button transition:fly={{duration: 500, delay: 200, x:20}} class="absolute left-5 top-5" style:font-size="12px" on:click={() => {
        assetsVisible.set($assetsVisible ? 0 : 1);
      }}>show assets</button>
    {/if}
    <Header />
    <Workspace />
  </div>
  <div 
    class='
      surface outputs transition-colors border-l hover:border-gray-300 border-transparent
      fixed
    '
    style:padding-right="{($inspectorVisibilityTween * 80)}px"
    style:right="{$panes.right * (1 - $inspectorVisibilityTween)}px" 
  >
    {$inspectorVisibilityTween}
    <InspectorSidebar />
    <button class="fixed" style:right="{$panes.right - 20}px" on:click={() => { 
      inspectorVisible.set($inspectorVisible ? 0 : 1); 
      }}>
        HIDE
    </button>
  </div>
  <div
    style:display="none"
    class='preview-drawer bg-white'
    style:height="var(--bottom-sidebar-width, 300px)"
    style:grid-area="preview" 
    style:align-self="end">
      <PreviewDrawer />
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

.surface.outputs, .surface.assets {
  overflow-y: auto;
  overflow-x: hidden;
}

.preview-drawer {
  overflow: hidden;
}

</style>