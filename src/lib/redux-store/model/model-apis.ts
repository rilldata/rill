import { dataModelerService } from "$lib/application-state-stores/application-store";
import { selectApplicationActiveEntity } from "$lib/redux-store/application/application-selectors";
import { selectMetricsDefinitionsByModelId } from "$lib/redux-store/dimension-definition/dimension-definition-selectors";
import { validateSelectedSources } from "$lib/redux-store/metrics-definition/metrics-definition-apis";
import { store } from "$lib/redux-store/store-root";
import { queryClient } from "$lib/svelte-query/globalQueryClient";
import { invalidateMetricsView } from "$lib/svelte-query/queries/metrics-view";

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

  syncMetricsDefinitions(modelId);
};

export const deleteModelApi = async (modelId: string) => {
  await dataModelerService.dispatch("deleteModel", [modelId]);
  syncMetricsDefinitions(modelId);
};

const syncMetricsDefinitions = (modelId: string) => {
  const state = store.getState();
  const activeEntity = selectApplicationActiveEntity(state);

  const metricsDefinitions = selectMetricsDefinitionsByModelId(state, modelId);
  metricsDefinitions.forEach((metricsDefinition) => {
    invalidateMetricsView(queryClient, metricsDefinition.id);
    if (activeEntity.id === metricsDefinition.id) {
      store.dispatch(validateSelectedSources(metricsDefinition.id));
    }
  });
};
