<script>
import { getContext } from "svelte";

import Workspace from "./_surfaces/workspace/index.svelte";
import InspectorSidebar from "./_surfaces/inspector/index.svelte";
import AssetsSidebar from "./_surfaces/assets/index.svelte";
import PreviewDrawer from "./_surfaces/preview/index.svelte";

const store = getContext("rill:app:store");


</script>

<div class="surface assets" style:grid-area="left-pane">
  <AssetsSidebar />
</div>
<div class="surface inputs" style:grid-area="workspace">
  <Workspace />
</div>
<div class='surface outputs  border-l border-gray-300' style:grid-area="right-pane">
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

<style>
.inputs {
  --hue: 217;
  --sat: 20%;
  --lgt: 95%;
  --bg: hsl(var(--hue), var(--sat), var(--lgt));
  --bg-transparent: hsla(var(--hue), var(--sat), var(--lgt), .8);
  background-color: var(--bg);
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