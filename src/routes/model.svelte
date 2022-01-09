<script>
import { getContext } from "svelte";
import { fade } from "svelte/transition";

import AddIcon from "$lib/components/icons/AddIcon.svelte";
import RefreshIcon from "$lib/components/icons/RefreshIcon.svelte";
import Logo from "$lib/components/Logo.svelte";
import Spinner from "$lib/components/Spinner.svelte";

import EditorPane from "./_panes/EditorPane.svelte";
import InspectorPane from "./_panes/InspectorPane.svelte";
import AssetsPane from "./_panes/AssetsPane.svelte";

let resultset;
let queryInfo;
let query;
let destinationInfo;

// FIXME: this is out of control :(
let destinationSize;

const store = getContext("rill:app:store");

let dbRunState = 'disconnected';
let runstateTimer;

function debounceRunstate(state) {
  if (runstateTimer) clearTimeout(runstateTimer);
  setTimeout(() => {
    dbRunState = state;
  }, 500)
}

$: debounceRunstate($store?.status || 'disconnected');

</script>

<header class="header pr-3">
  <h1><Logo /></h1>
  <button  on:click={() => store.action("addQuery", {})}><AddIcon size={18} /></button>
  <button on:click={() => store.action('reset')}>
      <RefreshIcon size={18} />
  </button>
  <div></div>
  <div class="self-center">
    {#if dbRunState === 'running'}
      <div transition:fade={{ duration: 300 }}>
        <Spinner />
      </div>
    {/if}
  </div>
</header>
<div class='body'>
  <div class="pane assets">
    <AssetsPane />
  </div>
  <div class="pane inputs">
    <EditorPane />
  </div>

  <div class='pane outputs'>
    <InspectorPane />
    </div>
  </div>

<style>
.body {
  width: calc(100vw);
  display: grid;
  grid-template-columns: max-content auto max-content;
  align-content: stretch;
  min-height: calc(100vh - var(--header-height));
}

header {
  box-sizing: border-box;
  margin:0;
  background: linear-gradient(to right, hsl(300, 30%, 14%), hsl(300, 60%, 18%));
  color: white;
  height: var(--header-height);
  display: grid;
  justify-items: left;
  justify-content: stretch;
  align-items: stretch;
  align-content: stretch;
  grid-template-columns: max-content max-content max-content auto max-content;
}

header h1 {
  font-size:13px;
  font-weight: normal;
  margin:0;
  padding:0;
  display: grid;
  place-items: center;
  padding: 0px 12px;
  padding-left: 2px;
  margin-left: 1rem;
}

header button {
  color: white;
  background-color: transparent;
  display: grid;
  place-items: center;
  padding: 0px 12px;
  border:none;
  font-size: 1.5rem;
}

header button:hover {
  background-color: hsla(var(--hue), var(--sat), var(--lgt), .1);
}

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

.outputs {
  /* padding: 1rem; */
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