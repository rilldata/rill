import { generateFilteredEntitySelectors } from "$lib/redux-store/utils/selector-utils";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";

export const {
  singleSelector: selectDimensionById,
  manySelector: selectDimensionsByMetricsId,
} = generateFilteredEntitySelectors<[string], DimensionDefinitionEntity>(
  "dimensionDefinition",
  (entity: DimensionDefinitionEntity, metricsDefId: string) =>
    entity.metricsDefId === metricsDefId
);
