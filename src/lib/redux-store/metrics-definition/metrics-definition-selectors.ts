import { generateEntitySelectors } from "$lib/redux-store/utils/selector-utils";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

export const {
  manySelector: selectAllMetricsDefinitions,
  singleSelector: selectMetricsDefinitionById,
} = generateEntitySelectors<MetricsDefinitionEntity>("metricsDefinition");
