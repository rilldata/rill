import { generateFilteredEntitySelectors } from "../utils/selector-utils";
import type { DimensionDefinitionEntity } from "$web-local/common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { RillReduxState } from "../store-root";
import { ValidationState } from "$web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

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

export const dimensionIsValid = (dimension: DimensionDefinitionEntity) =>
  dimension.dimensionIsValid === ValidationState.OK;
export const selectValidDimensions = (
  dimensions: Array<DimensionDefinitionEntity>
) => dimensions.filter(dimensionIsValid);

export const selectValidDimensionsByMetricsId = (
  state: RillReduxState,
  metricsDefId: string
) => selectValidDimensions(selectDimensionsByMetricsId(state, metricsDefId));
