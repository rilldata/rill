import { dataModelerService } from "../../application-state-stores/application-store";
import { selectApplicationActiveEntity } from "../application/application-selectors";
import { selectMetricsDefinitionsByModelId } from "../dimension-definition/dimension-definition-selectors";
import { validateSelectedSources } from "../metrics-definition/metrics-definition-apis";
import { store } from "../store-root";
import { queryClient } from "../../svelte-query/globalQueryClient";
import { invalidateMetricsView } from "../../svelte-query/queries/metrics-views/invalidation";

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
