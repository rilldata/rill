import { generateFilteredSelectors } from "$lib/redux-store/slice-utils";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";

export const {
  singleSelector: selectMeasureById,
  manySelector: selectMeasuresByMetricsId,
} = generateFilteredSelectors(
  "measureDefinition",
  (entity: MeasureDefinitionEntity, metricsDefId: string) =>
    entity.metricsDefId === metricsDefId
);
