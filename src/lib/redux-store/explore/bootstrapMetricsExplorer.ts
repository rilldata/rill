import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { bootstrapMetricsDefinition } from "$lib/redux-store/metrics-definition/bootstrapMetricsDefinition";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { selectValidDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-selectors";
import { selectValidMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-selectors";
import { dataModelerService } from "$lib/application-state-stores/application-store";
import { syncExplore } from "$lib/redux-store/explore/explore-apis";
import { selectMetricsExplorerById } from "$lib/redux-store/explore/explore-selectors";

export const bootstrapMetricsExplorer = createAsyncThunk(
  `${EntityType.MetricsExplorer}/bootstrapMetricsExplorer`,
  async (metricsDefId: string, thunkAPI) => {
    // fetch all entities related to the metrics definition
    await thunkAPI.dispatch(bootstrapMetricsDefinition(metricsDefId));

    const state = thunkAPI.getState() as RillReduxState;
    const dimensions = selectValidDimensionsByMetricsId(state, metricsDefId);
    const measures = selectValidMeasuresByMetricsId(state, metricsDefId);
    // if there are no valid dimensions or measures take the user to metrics definition editor
    if (!dimensions.length || !measures.length) {
      await dataModelerService.dispatch("setActiveAsset", [
        EntityType.MetricsDefinition,
        metricsDefId,
      ]);
      return;
    }

    await syncExplore(
      thunkAPI.dispatch,
      metricsDefId,
      selectMetricsExplorerById(state, metricsDefId),
      dimensions,
      measures
    );
  }
);
