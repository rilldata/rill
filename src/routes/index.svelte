<script>
import { getContext } from "svelte";

import Workspace from "./_surfaces/Workspace.svelte";
import InspectorSidebar from "./_surfaces/InspectorSidebar.svelte";
import AssetsSidebar from "./_surfaces/AssetsSidebar.svelte";

const store = getContext("rill:app:store");


</script>

<div class="surface assets">
  <AssetsSidebar />
</div>
<div class="surface inputs">
  <Workspace />
</div>
<div class='surface outputs'>
  <InspectorSidebar />
</div>

<style>
.inputs {
  --hue: 217;
  --sat: 20%;
  --lgt: 95%;
  --bg: hsl(var(--hue), var(--sat), var(--lgt));
  --bg-transparent: hsla(var(--hue), var(--sat), var(--lgt), .8);
  background-color: var(--bg);
  height: calc(100vh - var(--header-height));
  overflow-y: auto;
}

.surface {
  box-sizing: border-box;
}

.surface:first-child {
  border-right: 1px solid #ddd;
}

.surface.outputs, .surface.assets {
  height: calc(100vh - var(--header-height));
  overflow-y: auto;
  overflow-x: hidden;
}

</style>