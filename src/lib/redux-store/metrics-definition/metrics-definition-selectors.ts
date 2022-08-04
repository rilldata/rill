import { generateEntitySelectors } from "$lib/redux-store/utils/selector-utils";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { RillReduxState } from "$lib/redux-store/store-root";

export const {
  manySelector: selectAllMetricsDefinitions,
  singleSelector: selectMetricsDefinitionById,
} = generateEntitySelectors<MetricsDefinitionEntity, "metricsDefinition">(
  "metricsDefinition"
);

export const selectMetricsDefinitionMatchingName = (
  state: RillReduxState,
  name: string
) => {
  return state.metricsDefinition.ids
    .filter((metricsDefId) =>
      state.metricsDefinition.entities[metricsDefId].metricDefLabel.includes(
        name
      )
    )
    .map((metricsDefId) => state.metricsDefinition.entities[metricsDefId]);
};
