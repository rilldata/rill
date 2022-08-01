import { dataModelerService } from "$lib/application-state-stores/application-store";
import { selectMetricsDefinitionsByModelId } from "$lib/redux-store/dimension-definition/dimension-definition-selectors";
import { store } from "$lib/redux-store/store-root";
import { setExplorerIsStale } from "$lib/redux-store/explore/explore-slice";

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

  const metricsDefinitions = selectMetricsDefinitionsByModelId(
    store.getState(),
    modelId
  );
  metricsDefinitions.forEach((metricsDefinition) =>
    store.dispatch(setExplorerIsStale(metricsDefinition.id, true))
  );
};
