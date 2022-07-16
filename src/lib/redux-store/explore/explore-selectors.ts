import { generateEntitySelectors } from "$lib/redux-store/utils/selector-utils";
import type { MetricsExploreEntity } from "$lib/redux-store/explore/explore-slice";
import type { RillReduxState } from "$lib/redux-store/store-root";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import { prune } from "../../../routes/_surfaces/workspace/explore/utils";
import { selectMeasureById } from "$lib/redux-store/measure-definition/measure-definition-selectors";
import type { BasicMeasureDefinition } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";

export const {
  manySelector: selectMetricsExplores,
  singleSelector: selectMetricsExploreById,
} = generateEntitySelectors<MetricsExploreEntity, "metricsLeaderboard">(
  "metricsLeaderboard"
);

/**
 * Common code to fetch metrics explore and some params from state
 * Fetches 'measures' from state if not passed.
 * Fetches 'filters' from state if not passed. Also prunes empty filters.
 */
export const selectMetricsExploreParams = (
  state: RillReduxState,
  id: string,
  {
    measures,
    filters,
    dimensions,
  }: {
    measures?: Array<MeasureDefinitionEntity>;
    filters?: ActiveValues;
    dimensions: Record<string, DimensionDefinitionEntity>;
  }
) => {
  const metricsExplore = selectMetricsExploreById(state, id);

  if (!filters) {
    filters = metricsExplore.activeValues;
  }
  filters = prune(filters, dimensions);

  if (!measures) {
    measures = metricsExplore.selectedMeasureIds.map((measureId) =>
      selectMeasureById(state, measureId)
    );
  }

  return {
    metricsExplore,
    prunedFilters: filters,
    normalisedMeasures: measures.map(
      (measure) =>
        ({
          id: measure.id,
          expression: measure.expression,
          sqlName: measure.sqlName,
        } as BasicMeasureDefinition)
    ),
  };
};
