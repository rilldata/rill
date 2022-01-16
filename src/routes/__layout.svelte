<script>
import "../fonts.css";
import "../app.css";
import { setContext } from "svelte";
import { createStore } from '$lib/app-store';
import { browser } from "$app/env";

import { fade } from "svelte/transition";

import AddIcon from "$lib/components/icons/AddIcon.svelte";
import RefreshIcon from "$lib/components/icons/RefreshIcon.svelte";
import Spinner from "$lib/components/Spinner.svelte";
import Logo from "$lib/components/Logo.svelte";
import NotificationCenter from "$lib/components/notifications/NotificationCenter.svelte";
import notification from "$lib/components/notifications/";

let store;

if (browser) {
  store = createStore();
  setContext('rill:app:store', store);
  notification.listenToSocket(store.socket);
}


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
  <div class='grid grid-flow-col'>
    <div id="controls" class="grid grid-flow-col">
      <!-- FIXME: move this to slot in __layout -->
      <button  on:click={() => store.action("addQuery", {})}><AddIcon size={18} /></button>
      <button on:click={() => store.action('reset')}>
          <RefreshIcon size={18} />
      </button>
    </div>
  </div>
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
  <slot />
  </div>

<NotificationCenter />


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
    grid-template-columns: max-content max-content auto max-content;
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
  
  </style>