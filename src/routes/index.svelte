<script>
import { getContext } from "svelte";

import Workspace from "./_surfaces/workspace/index.svelte";
import InspectorSidebar from "./_surfaces/inspector/index.svelte";
import AssetsSidebar from "./_surfaces/assets/index.svelte";
import PreviewDrawer from "./_surfaces/preview/index.svelte";
import Header from "./_surfaces/header/index.svelte";
import { tweened } from "svelte/motion";
import { cubicOut } from "svelte/easing";

const store = getContext("rill:app:store");
</script>

<div class='body'>
  <div class="surface assets" style:grid-area="left-pane">
    <AssetsSidebar />
  </div>
  <div style:grid-area="header-bar">
    <Header />
  </div>
  <div class="surface inputs bg-gray-100" style:grid-area="workspace">
    <Workspace />
  </div>
  <div class='surface outputs transition-colors border-l hover:border-gray-300 border-transparent' style:grid-area="right-pane">
    <InspectorSidebar />
  </div>
  <div
    style:display=none
    class='preview-drawer bg-white'
    style:height="var(--bottom-sidebar-width, 0px)"
    style:grid-area="preview" 
    style:align-self="end">
      <PreviewDrawer />
  </div>
</div>
<style>

.body {
    width: 100vw;
    display: grid;
    grid-template-columns: max-content auto var(--right-sidebar-width, 400px);
    grid-template-rows: var(--header-height, 60px) auto var(--preview-height, 0px);
    grid-template-areas: "left-pane header-bar right-pane"
                          "left-pane workspace right-pane"
                         "left-pane preview preview";
    align-content: stretch;
    align-items: stretch;
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