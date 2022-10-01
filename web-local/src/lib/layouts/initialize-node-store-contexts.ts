import { browser } from "$app/environment";
import {
  createDerivedModelStore,
  createPersistentModelStore,
} from "@rilldata/web-local/lib/application-state-stores/model-stores";
import { createQueryHighlightStore } from "@rilldata/web-local/lib/application-state-stores/query-highlight-store";
import notificationStore from "@rilldata/web-local/lib/components/notifications/";
import type { ApplicationMetadata } from "@rilldata/web-local/lib/types";
import { setContext } from "svelte";
import { createStore } from "../application-state-stores/application-store";
import {
  createDerivedTableStore,
  createPersistentTableStore,
} from "../application-state-stores/table-stores";
import { syncApplicationState } from "../redux-store/application/application-apis";

/** This function will initialize the existing node stores and will connect them
 * to the Node server.
 */
export function initializeNodeStoreContexts() {
  let store;
  const queryHighlight = createQueryHighlightStore();

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
    notificationStore.listenToSocket(store.socket);
    syncApplicationState(store);
  }
}
