import { generateBasicSelectors } from "$lib/redux-store/slice-utils";

export const {
  manySelector: selectAllMetricsDefinitions,
  singleSelector: selectMetricsDefinitionById,
} = generateBasicSelectors("metricsDefinition");
