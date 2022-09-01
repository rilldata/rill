import { dataModelerService } from "$lib/application-state-stores/application-store";
import { selectMetricsDefinitionsByModelId } from "$lib/redux-store/dimension-definition/dimension-definition-selectors";
import { store } from "$lib/redux-store/store-root";
import { selectApplicationActiveEntity } from "$lib/redux-store/application/application-selectors";
import { validateSelectedSources } from "$lib/redux-store/metrics-definition/metrics-definition-apis";
import { invalidateMetricView } from "$lib/svelte-query/queries/metric-view";
import { queryClient } from "$lib/svelte-query/globalQueryClient";

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
    invalidateMetricView(queryClient, metricsDefinition.id);
    if (activeEntity.id === metricsDefinition.id) {
      store.dispatch(validateSelectedSources(metricsDefinition.id));
    }
  });
};
