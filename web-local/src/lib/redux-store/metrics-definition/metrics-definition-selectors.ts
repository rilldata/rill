import { getName } from "@rilldata/web-local/common/utils/incrementName";
import { generateEntitySelectors } from "../utils/selector-utils";
import type { MetricsDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { RillReduxState } from "../store-root";

export const {
  manySelector: selectAllMetricsDefinitions,
  singleSelector: selectMetricsDefinitionById,
} = generateEntitySelectors<MetricsDefinitionEntity, "metricsDefinition">(
  "metricsDefinition"
);

export const selectNextMetricsDefinitionName = (
  state: RillReduxState,
  name: string
) => {
  return getName(
    name,
    state.metricsDefinition.ids.map(
      (metricsDefId) =>
        state.metricsDefinition.entities[metricsDefId].metricDefLabel
    )
  );
};
