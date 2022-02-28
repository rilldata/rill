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

import { createQueryHighlightStore } from "$lib/query-highlight-store";

let store;
let queryHighlight = createQueryHighlightStore();
if (browser) {
  store = createStore();
  setContext('rill:app:store', store);
  setContext('rill:app:query-highlight', queryHighlight);
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

<div class='body'>
  <slot />
  </div>

<NotificationCenter />