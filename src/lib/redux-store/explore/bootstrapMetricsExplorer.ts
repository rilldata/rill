import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { bootstrapMetricsDefinition } from "$lib/redux-store/metrics-definition/bootstrapMetricsDefinition";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { selectMetricsExplorerById } from "$lib/redux-store/explore/explore-selectors";
import { syncExplore } from "$lib/redux-store/explore/explore-apis";
import { selectDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-selectors";
import { selectMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-selectors";

export const bootstrapMetricsExplorer = createAsyncThunk(
  `${EntityType.MetricsExplorer}/bootstrapMetricsExplorer`,
  async (metricsDefId: string, thunkAPI) => {
    // fetch all entities related to the metrics definition
    await thunkAPI.dispatch(bootstrapMetricsDefinition(metricsDefId));

    const state = thunkAPI.getState() as RillReduxState;
    await syncExplore(
      thunkAPI.dispatch,
      metricsDefId,
      selectMetricsExplorerById(state, metricsDefId),
      selectDimensionsByMetricsId(state, metricsDefId),
      selectMeasuresByMetricsId(state, metricsDefId)
    );
  }
);
