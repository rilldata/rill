import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import { getArrayDiff } from "$common/utils/getArrayDiff";
import { setReferenceValues } from "$lib/redux-store/big-number/big-number-slice";
import {
  addDimensionToExplore,
  addMeasureToExplore,
  clearSelectedLeaderboardValues,
  initMetricsExplorer,
  MetricsExplorerEntity,
  removeDimensionFromExplore,
  removeMeasureFromExplore,
  setExplorerIsStale,
  setExploreSelectedTimeRange,
  setLeaderboardMeasureId,
} from "$lib/redux-store/explore/explore-slice";
import { selectValidMeasures } from "$lib/redux-store/measure-definition/measure-definition-selectors";

/**
 * Syncs explore with updated measures and dimensions.
 * If a MetricsExplorer entity is not present then a new one is created.
 */
export const syncExplore = async (
  dispatch,
  metricsDefId: string,
  metricsExplorer: MetricsExplorerEntity,
  dimensions: Array<DimensionDefinitionEntity>,
  measures: Array<MeasureDefinitionEntity>
) => {
  if (measures) measures = selectValidMeasures(measures);

  let shouldUpdate = false;
  if (!metricsExplorer) {
    dispatch(initMetricsExplorer(metricsDefId, dimensions, measures));
    shouldUpdate = true;
  } else {
    if (dimensions)
      shouldUpdate = syncDimensions(dispatch, metricsExplorer, dimensions);
    if (measures)
      shouldUpdate ||= syncMeasures(dispatch, metricsExplorer, measures);
  }

  // To avoid infinite loop only update if something changed.
  if (shouldUpdate || metricsExplorer.isStale) {
    dispatch(setExplorerIsStale(metricsDefId, false));
  }
};
/**
 * Syncs dimensions from MetricsDefinition with MetricsExplorer dimensions.
 * Calls {@link addDimensionToExplore} for missing dimension.
 * Calls {@link removeDimensionFromExplore} for excess dimension.
 */
const syncDimensions = (
  dispatch,
  metricsExplorer: MetricsExplorerEntity,
  dimensions: Array<DimensionDefinitionEntity>
) => {
  const { extraSrc: addDimensions, extraTarget: removeDimensions } =
    getArrayDiff(
      dimensions,
      (dimension) => dimension.id,
      metricsExplorer.leaderboards,
      (leaderboard) => leaderboard.dimensionId
    );
  addDimensions.forEach((addDimension) =>
    dispatch(addDimensionToExplore(metricsExplorer.id, addDimension.id))
  );
  removeDimensions.forEach((removeDimension) =>
    dispatch(
      removeDimensionFromExplore(
        metricsExplorer.id,
        removeDimension.dimensionId
      )
    )
  );

  return addDimensions.length > 0 || removeDimensions.length > 0;
};
/**
 * Syncs measures from MetricsDefinition with MetricsExplorer measures.
 * Calls {@link addMeasureToExplore} for missing measure.
 * Calls {@link removeMeasureFromExplore} for excess measure.
 */
const syncMeasures = (
  dispatch,
  metricsExplorer: MetricsExplorerEntity,
  measures: Array<MeasureDefinitionEntity>
) => {
  const { extraSrc: addMeasures, extraTarget: removeMeasures } = getArrayDiff(
    measures,
    (measure) => measure.id,
    metricsExplorer.measureIds,
    (measureId) => measureId
  );
  addMeasures.forEach((addMeasure) =>
    dispatch(addMeasureToExplore(metricsExplorer.id, addMeasure.id))
  );
  removeMeasures.forEach((removeMeasure) =>
    dispatch(removeMeasureFromExplore(metricsExplorer.id, removeMeasure))
  );

  return addMeasures.length > 0 || removeMeasures.length > 0;
};

/**
 * Sets the measure id used in leaderboard for ranking and other calculations.
 * It then updates Leaderboard values based on selected filters and measure
 */
export const setMeasureIdAndUpdateLeaderboard = (
  dispatch,
  metricsDefId: string,
  measureId: string
) => {
  dispatch(setLeaderboardMeasureId(metricsDefId, measureId));
};

/**
 * Clears all selected values in the leaderboard.
 */
export const clearSelectedLeaderboardValuesAndUpdate = (
  dispatch,
  metricsDefId: string
) => {
  dispatch(clearSelectedLeaderboardValues(metricsDefId));
};

/**
 * Sets user selected time range.
 */
export const setExploreSelectedTimeRangeAndUpdate = (
  dispatch,
  metricsDefId: string,
  timeRange: TimeSeriesTimeRange
) => {
  dispatch(setReferenceValues(metricsDefId, undefined));
  dispatch(setExploreSelectedTimeRange(metricsDefId, timeRange));
};
