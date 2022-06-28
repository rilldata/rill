import { generateBasicSelectors } from "$lib/redux-store/utils/selector-utils";

export const {
  manySelector: selectAllMetricsDefinitions,
  singleSelector: selectMetricsDefinitionById,
} = generateBasicSelectors("metricsDefinition");
