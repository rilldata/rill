import { generateFilteredEntitySelectors } from "$lib/redux-store/utils/selector-utils";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { RillReduxState } from "$lib/redux-store/store-root";

export const {
  singleSelector: selectDimensionById,
  manySelector: selectDimensionsByMetricsId,
} = generateFilteredEntitySelectors<
  [string],
  DimensionDefinitionEntity,
  "dimensionDefinition"
>(
  "dimensionDefinition",
  (entity: DimensionDefinitionEntity, metricsDefId: string) =>
    entity.metricsDefId === metricsDefId
);

export const selectMetricsDefinitionsByModelId = (
  state: RillReduxState,
  modelId: string
) =>
  state.metricsDefinition.ids
    .map((id) => state.metricsDefinition.entities[id])
    .filter((metricsDefinition) => metricsDefinition.sourceModelId === modelId);
