<script>
import "../app.css";
import { setContext } from "svelte";
import { createStore } from '$lib/app-store';
import { browser } from "$app/env";

import NotificationCenter from "$lib/components/notifications/NotificationCenter.svelte";
import notification from "$lib/components/notifications/";

let store;

if (browser) {
  store = createStore();
  setContext('rill:app:store', store);
  notification.listenToSocket(store.socket);
}

</script>

<slot />

<NotificationCenter />