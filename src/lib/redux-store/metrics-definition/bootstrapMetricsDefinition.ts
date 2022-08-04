import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { selectMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-selectors";
import { selectDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-selectors";
import { fetchManyMeasuresApi } from "$lib/redux-store/measure-definition/measure-definition-apis";
import { fetchManyDimensionsApi } from "$lib/redux-store/dimension-definition/dimension-definition-apis";
import { validateSelectedSources } from "$lib/redux-store/metrics-definition/metrics-definition-apis";

/**
 * Bootstrap a metrics definition.
 * 1. If an entity is not present fetch all metrics definitions.
 */
export const bootstrapMetricsDefinition = createAsyncThunk(
  `${EntityType.MeasureDefinition}/bootstrapMetricsDefinition`,
  async (metricsDefId: string, thunkAPI) => {
    const state = thunkAPI.getState() as RillReduxState;

    const measures = selectMeasuresByMetricsId(state, metricsDefId);
    const dimensions = selectDimensionsByMetricsId(state, metricsDefId);
    await Promise.all([
      !measures.length
        ? thunkAPI.dispatch(fetchManyMeasuresApi({ metricsDefId }))
        : Promise.resolve(),
      !dimensions.length
        ? thunkAPI.dispatch(fetchManyDimensionsApi({ metricsDefId }))
        : Promise.resolve(),
    ]);
    thunkAPI.dispatch(validateSelectedSources(metricsDefId));
  }
);
