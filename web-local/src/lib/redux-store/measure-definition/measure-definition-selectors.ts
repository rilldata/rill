import { generateFilteredEntitySelectors } from "../utils/selector-utils";
import type { MeasureDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { getFallbackMeasureName } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { RillReduxState } from "../store-root";
import { ValidationState } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

export const {
  singleSelector: selectMeasureById,
  manySelector: selectMeasuresByMetricsId,
  manySelectorByIds: selectMeasuresByIds,
} = generateFilteredEntitySelectors<
  [string],
  MeasureDefinitionEntity,
  "measureDefinition"
>(
  "measureDefinition",
  (entity: MeasureDefinitionEntity, metricsDefId: string) =>
    entity.metricsDefId === metricsDefId
);

export const selectMeasureFieldNameByIdAndIndex = (
  store: RillReduxState,
  id: string,
  index: number
) => {
  const measure = selectMeasureById(store, id);
  return measure ? getFallbackMeasureName(index, measure.sqlName) : "";
};

export const measureIsValid = (measure: MeasureDefinitionEntity) =>
  measure.expressionIsValid === ValidationState.OK;
export const selectValidMeasures = (measures: Array<MeasureDefinitionEntity>) =>
  measures.filter(measureIsValid);

export const selectValidMeasuresByMetricsId = (
  state: RillReduxState,
  metricsDefId: string
) => selectValidMeasures(selectMeasuresByMetricsId(state, metricsDefId));
