<script lang="ts">
  import { browser } from "$app/env";
  import { createStore } from "$lib/application-state-stores/application-store";
  import {
    createDerivedModelStore,
    createPersistentModelStore,
  } from "$lib/application-state-stores/model-stores";
  import { createQueryHighlightStore } from "$lib/application-state-stores/query-highlight-store";
  import {
    createDerivedTableStore,
    createPersistentTableStore,
  } from "$lib/application-state-stores/table-stores";
  import notification from "$lib/components/notifications/";
  import NotificationCenter from "$lib/components/notifications/NotificationCenter.svelte";
  import { initMetrics } from "$lib/metrics/initMetrics";
  import { syncApplicationState } from "$lib/redux-store/application/application-apis";
  import type { ApplicationMetadata } from "$lib/types";
  import { onMount, setContext } from "svelte";
  import "../app.css";
  import "../fonts.css";
  import { QueryClient, QueryClientProvider } from "@sveltestack/svelte-query";

  let store;
  let queryHighlight = createQueryHighlightStore();

  const applicationMetadata: ApplicationMetadata = {
    version: RILL_VERSION, // constant defined in svelte.config.js
    commitHash: RILL_COMMIT, // constant defined in svelte.config.js
  };

  setContext("rill:app:metadata", applicationMetadata);

  if (browser) {
    store = createStore();
    setContext("rill:app:store", store);
    setContext("rill:app:query-highlight", queryHighlight);
    setContext(`rill:app:persistent-table-store`, createPersistentTableStore());
    setContext(`rill:app:derived-table-store`, createDerivedTableStore());
    setContext(`rill:app:persistent-model-store`, createPersistentModelStore());
    setContext(`rill:app:derived-model-store`, createDerivedModelStore());
    notification.listenToSocket(store.socket);
    syncApplicationState(store);
  }

  onMount(() => {
    initMetrics();
  });

  let dbRunState = "disconnected";
  let runstateTimer;

  function debounceRunstate(state) {
    if (runstateTimer) clearTimeout(runstateTimer);
    setTimeout(() => {
      dbRunState = state;
    }, 500);
  }

  $: debounceRunstate($store?.status || "disconnected");

  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        refetchOnWindowFocus: false,
        placeholderData: [],
      },
    },
  });
</script>

<QueryClientProvider client={queryClient}>
  <div class="body">
    <slot />
  </div>
</QueryClientProvider>

<NotificationCenter />
