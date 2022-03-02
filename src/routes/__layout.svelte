<script>
import "../fonts.css";
import "../app.css";
import { setContext } from "svelte";
import { createStore } from '$lib/app-store';
import { browser } from "$app/env";

import NotificationCenter from "$lib/components/notifications/NotificationCenter.svelte";
import notification from "$lib/components/notifications/";

import { createQueryHighlightStore } from "$lib/query-highlight-store";
import { createDerivedTableStore, createPersistentTableStore } from "$lib/tableStores.ts";
import { createDerivedModelStore, createPersistentModelStore } from "$lib/modelStores.ts";

let store;
let queryHighlight = createQueryHighlightStore();
if (browser) {
  store = createStore();
  setContext('rill:app:store', store);
  setContext('rill:app:query-highlight', queryHighlight);
  setContext(`rill:app:persistent-table-store`, createPersistentTableStore());
  setContext(`rill:app:derived-table-store`, createDerivedTableStore());
  setContext(`rill:app:persistent-model-store`, createPersistentModelStore());
  setContext(`rill:app:derived-model-store`, createDerivedModelStore());
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
