import { dataModelerService } from "$lib/application-state-stores/application-store";
import { selectMetricsDefinitionsByModelId } from "$lib/redux-store/dimension-definition/dimension-definition-selectors";
import { store } from "$lib/redux-store/store-root";
import { setExplorerIsStale } from "$lib/redux-store/explore/explore-slice";
import { selectApplicationActiveEntity } from "$lib/redux-store/application/application-selectors";
import { validateSelectedSources } from "$lib/redux-store/metrics-definition/metrics-definition-apis";

export const updateModelQueryApi = async (
  modelId: string,
  modelQuery: string,
  force = false
) => {
  await dataModelerService.dispatch("updateModelQuery", [
    modelId,
    modelQuery,
    force,
  ]);

  const state = store.getState();
  const activeEntity = selectApplicationActiveEntity(state);

  const metricsDefinitions = selectMetricsDefinitionsByModelId(state, modelId);
  metricsDefinitions.forEach((metricsDefinition) => {
    store.dispatch(setExplorerIsStale(metricsDefinition.id, true));
    if (activeEntity.id === metricsDefinition.id) {
      store.dispatch(validateSelectedSources(metricsDefinition.id));
    }
  });
};
