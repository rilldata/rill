import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { prune } from "../../../routes/_surfaces/workspace/explore/utils";
import { fetchWrapper, streamingFetchWrapper } from "$lib/util/fetchWrapper";
import {
  addDimensionToExplore,
  addMeasureToExplore,
  clearSelectedLeaderboardValues,
  initMetricsExplore,
  LeaderboardValues,
  MetricsExploreEntity,
  setExploreSelectedTimeRange,
  setExploreTimeRange,
  removeDimensionFromExplore,
  removeMeasureFromExplore,
  setLeaderboardDimensionValues,
  setLeaderboardMeasureId,
  toggleExploreMeasure,
  toggleLeaderboardActiveValue,
} from "$lib/redux-store/explore/explore-slice";
import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import { generateTimeSeriesApi } from "$lib/redux-store/timeseries/timeseries-apis";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { generateBigNumbersApi } from "$lib/redux-store/big-number/big-number-apis";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import { getArrayDiff } from "$common/utils/getArrayDiff";

/**
 * A wrapper to dispatch updates to explore.
 * Currently, it updates these sections
 * 1. Leaderboard values based on selected filters and measure
 * 2. Time series for all selected measures
 * 3. Big numbers for all selected measures
 */
const updateExploreWrapper = (dispatch, metricsDefId: string) => {
  dispatch(generateTimeSeriesApi({ id: metricsDefId }));
  dispatch(generateBigNumbersApi({ id: metricsDefId }));
  dispatch(updateLeaderboardValuesApi(metricsDefId));
};

/**
 * Syncs explore with updated measures and dimensions.
 * If a MetricsExplore entity is not present then a new one is created.
 * It then calls {@link updateExploreWrapper} to update explore.
 * It also dispatches {@link fetchTimestampColumnRangeApi} to update time range.
 */
export const syncExplore = (
  dispatch,
  metricsDefId: string,
  metricsExplore: MetricsExploreEntity,
  dimensions: Array<DimensionDefinitionEntity>,
  measures: Array<MeasureDefinitionEntity>
) => {
  if (!metricsExplore) {
    dispatch(initMetricsExplore(metricsDefId, dimensions, measures));
    return;
  }

  let shouldUpdate = syncDimensions(dispatch, metricsExplore, dimensions);
  shouldUpdate ||= syncMeasures(dispatch, metricsExplore, measures);
  // To avoid infinite loop only update if something changed.
  // TODO: handle edge cases like measure expression or dimension column changing.
  if (shouldUpdate) {
    dispatch(fetchTimestampColumnRangeApi(metricsDefId));
    updateExploreWrapper(dispatch, metricsDefId);
  }
};
/**
 * Syncs dimensions from MetricsDefinition with MetricsExplore dimensions.
 * Calls {@link addDimensionToExplore} for missing dimension.
 * Calls {@link removeDimensionFromExplore} for excess dimension.
 */
const syncDimensions = (
  dispatch,
  metricsExplore: MetricsExploreEntity,
  dimensions: Array<DimensionDefinitionEntity>
) => {
  const { extraSrc: addDimensions, extraTarget: removeDimensions } =
    getArrayDiff(
      dimensions,
      (dimension) => dimension.id,
      metricsExplore.leaderboards,
      (leaderboard) => leaderboard.dimensionId
    );
  addDimensions.forEach((addDimension) =>
    dispatch(addDimensionToExplore(metricsExplore.id, addDimension.id))
  );
  removeDimensions.forEach((removeDimension) =>
    dispatch(
      removeDimensionFromExplore(metricsExplore.id, removeDimension.dimensionId)
    )
  );

  return addDimensions.length > 0 || removeDimensions.length > 0;
};
/**
 * Syncs measures from MetricsDefinition with MetricsExplore measures.
 * Calls {@link addMeasureToExplore} for missing measure.
 * Calls {@link removeMeasureFromExplore} for excess measure.
 */
const syncMeasures = (
  dispatch,
  metricsExplore: MetricsExploreEntity,
  measures: Array<MeasureDefinitionEntity>
) => {
  const { extraSrc: addMeasures, extraTarget: removeMeasures } = getArrayDiff(
    measures,
    (measure) => measure.id,
    metricsExplore.measureIds,
    (measureId) => measureId
  );
  addMeasures.forEach((addMeasure) =>
    dispatch(addMeasureToExplore(metricsExplore.id, addMeasure.id))
  );
  removeMeasures.forEach((removeMeasure) =>
    dispatch(removeMeasureFromExplore(metricsExplore.id, removeMeasure))
  );

  return addMeasures.length > 0 || removeMeasures.length > 0;
};

/**
 * Toggles selection of a measures to be displayed.
 * It then updates,
 * 1. Time series for all selected measures
 * 2. Big numbers for all selected measures
 */
export const toggleExploreMeasureAndUpdate = (
  dispatch,
  metricsDefId: string,
  selectedMeasureId: string
) => {
  dispatch(toggleExploreMeasure(metricsDefId, selectedMeasureId));
  dispatch(generateTimeSeriesApi({ id: metricsDefId }));
  dispatch(generateBigNumbersApi({ id: metricsDefId }));
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
  dispatch(updateLeaderboardValuesApi(metricsDefId));
};

/**
 * Toggles a selected value in the leaderboard.
 * Pass 'include' param boolean to denote whether the value is included or excluded in time series and big number queries.
 * It then calls {@link updateExploreWrapper} to update explore.
 */
export const toggleSelectedLeaderboardValueAndUpdate = (
  dispatch,
  metricsDefId: string,
  dimensionId: string,
  dimensionValue: unknown,
  include: boolean
) => {
  dispatch(
    toggleLeaderboardActiveValue(
      metricsDefId,
      dimensionId,
      dimensionValue,
      include
    )
  );
  updateExploreWrapper(dispatch, metricsDefId);
};

/**
 * Clears all selected values in the leaderboard.
 * It then calls {@link updateExploreWrapper} to update explore.
 */
export const clearSelectedLeaderboardValuesAndUpdate = (
  dispatch,
  metricsDefId: string
) => {
  dispatch(clearSelectedLeaderboardValues(metricsDefId));
  updateExploreWrapper(dispatch, metricsDefId);
};

/**
 * Sets user selected time rage.
 * It then calls {@link updateExploreWrapper} to update explore.
 */
export const setExploreSelectedTimeRangeAndUpdate = (
  dispatch,
  metricsDefId: string,
  selectedTimeRange: Partial<TimeSeriesTimeRange>
) => {
  dispatch(setExploreSelectedTimeRange(metricsDefId, selectedTimeRange));
  updateExploreWrapper(dispatch, metricsDefId);
};

/**
 * Async-thunk to update leaderboard values.
 * Streams dimension values from backend per dimension and updates it in the state.
 */
export const updateLeaderboardValuesApi = createAsyncThunk(
  `${EntityType.MetricsLeaderboard}/updateLeaderboard`,
  async (metricsDefId: string, thunkAPI) => {
    const state = thunkAPI.getState() as RillReduxState;
    const metricsExplore: MetricsExploreEntity = (
      thunkAPI.getState() as RillReduxState
    ).metricsLeaderboard.entities[metricsDefId];
    const filters = prune(
      metricsExplore.activeValues,
      state.dimensionDefinition.entities
    );
    const requestBody = {
      measureId: metricsExplore.leaderboardMeasureId,
      filters,
      timeRange: metricsExplore.selectedTimeRange,
    };

    const stream = streamingFetchWrapper<LeaderboardValues>(
      `metrics/${metricsExplore.id}/leaderboards`,
      "POST",
      requestBody
    );
    for await (const dimensionData of stream) {
      thunkAPI.dispatch(
        setLeaderboardDimensionValues(
          metricsExplore.id,
          dimensionData.dimensionId,
          dimensionData.values
        )
      );
    }
  }
);

/**
 * Fetches time range for the selected timestamp column.
 * Store the response in MetricsExplore slice by calling {@link setExploreTimeRange}
 */
export const fetchTimestampColumnRangeApi = createAsyncThunk(
  `${EntityType.MetricsLeaderboard}/getTimestampColumnRange`,
  async (metricsDefId: string, thunkAPI) => {
    const timeRange = await fetchWrapper(
      `metrics/${metricsDefId}/time-range`,
      "GET"
    );
    thunkAPI.dispatch(setExploreTimeRange(metricsDefId, timeRange));
  }
);
