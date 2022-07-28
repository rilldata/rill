import type {
  ActiveEntity,
  ApplicationState,
} from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type { ApplicationStore } from "$lib/application-state-stores/application-store";
import { get } from "svelte/store";
import { store } from "$lib/redux-store/store-root";
import { setApplicationActiveState } from "$lib/redux-store/application/application-slice";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { bootstrapMetricsDefinition } from "$lib/redux-store/metrics-definition/bootstrapMetricsDefinition";
import { bootstrapMetricsExplorer } from "$lib/redux-store/explore/bootstrapMetricsExplorer";

export const syncApplicationState = (appStore: ApplicationStore) => {
  appStore.subscribe(() => {
    const appState: ApplicationState = get(appStore);
    if (
      appState.activeEntity?.id !==
      store.getState().application.activeEntity?.id
    ) {
      store.dispatch(setApplicationActiveState(appState.activeEntity));
      activeEntityChangeApi(appState.activeEntity);
    }
  });
};

export const activeEntityChangeApi = (activeEntity: ActiveEntity) => {
  if (EntityType.MeasureDefinition) {
    store.dispatch(bootstrapMetricsDefinition(activeEntity.id));
  } else if (EntityType.MetricsExplorer) {
    store.dispatch(bootstrapMetricsExplorer(activeEntity.id));
  }
};
