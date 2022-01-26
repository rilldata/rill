<script>
import { getContext } from "svelte";

import Workspace from "./_panes/Workspace.svelte";
import InspectorPane from "./_panes/InspectorPane.svelte";
import AssetsPane from "./_panes/AssetsPane.svelte";

const store = getContext("rill:app:store");


</script>

<div class="pane assets">
  <AssetsPane />
</div>
<div class="pane inputs">
  <Workspace />
</div>
<div class='pane outputs'>
  <InspectorPane />
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

.pane {
  box-sizing: border-box;
}

.pane:first-child {
  border-right: 1px solid #ddd;
}

.pane.outputs, .pane.assets {
  height: calc(100vh - var(--header-height));
  overflow-y: auto;
  overflow-x: hidden;
}

</style>