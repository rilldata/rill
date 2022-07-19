import type { ApplicationState } from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type { ApplicationStore } from "$lib/application-state-stores/application-store";
import { get } from "svelte/store";
import { store } from "$lib/redux-store/store-root";
import { setApplicationActiveState } from "$lib/redux-store/application/application-slice";

export const syncApplicationState = (appStore: ApplicationStore) => {
  appStore.subscribe(() => {
    const appState: ApplicationState = get(appStore);
    if (
      appState.activeEntity?.id !==
      store.getState().application.activeEntity?.id
    ) {
      store.dispatch(setApplicationActiveState(appState.activeEntity));
    }
  });
};
