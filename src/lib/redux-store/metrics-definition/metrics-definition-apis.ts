import type { DerivedModelEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { SourceModelValidationStatus } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { asyncWait } from "$common/utils/waitUtils";
import { dataModelerService } from "$lib/application-state-stores/application-store";
import { selectApplicationActiveEntity } from "$lib/redux-store/application/application-selectors";
import {
  addOneDimension,
  clearDimensionsForMetricsDefId,
} from "$lib/redux-store/dimension-definition/dimension-definition-slice";
import {
  addOneMeasure,
  clearMeasuresForMetricsDefId,
} from "$lib/redux-store/measure-definition/measure-definition-slice";
import { selectMetricsDefinitionById } from "$lib/redux-store/metrics-definition/metrics-definition-selectors";
import {
  addManyMetricsDefs,
  addOneMetricsDef,
  removeMetricsDef,
  setSourceModelValidationStatus,
  setTimeDimensionValidationStatus,
  updateMetricsDef,
} from "$lib/redux-store/metrics-definition/metrics-definition-slice";
import { selectDerivedModelById } from "$lib/redux-store/model/model-selector";
import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import { RillReduxState, store } from "$lib/redux-store/store-root";
import { generateApis } from "$lib/redux-store/utils/api-utils";
import { streamingFetchWrapper } from "$lib/util/fetchWrapper";
import { invalidateExplorerThunk } from "$lib/redux-store/utils/invalidateExplorerThunk";
import { validateMeasureExpression } from "$lib/redux-store/measure-definition/measure-definition-apis";
import {
  selectMeasureById,
  selectMeasuresByMetricsId,
} from "$lib/redux-store/measure-definition/measure-definition-selectors";
import { validateDimensionColumnApi } from "$lib/redux-store/dimension-definition/dimension-definition-apis";
import { selectDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-selectors";

const handleMetricsDefDelete = async (id: string) => {
  const activeEntity = selectApplicationActiveEntity(store.getState());
  if (!activeEntity) return;

  if (
    activeEntity.id === id &&
    (activeEntity.type === EntityType.MetricsDefinition ||
      activeEntity.type === EntityType.MetricsExplorer)
  ) {
    const nextId = store.getState().metricsDefinition.ids[0];
    if (!nextId) {
      // TODO: refactor to use redux store once we move model and tables there.
      await dataModelerService.dispatch("setModelAsActiveAsset", []);
    } else {
      await dataModelerService.dispatch("setActiveAsset", [
        EntityType.MetricsDefinition,
        nextId as string,
      ]);
    }
  }
};

export const {
  fetchManyApi: fetchManyMetricsDefsApi,
  createApi: createMetricsDefsApi,
  updateApi: updateMetricsDefsApi,
  deleteApi: deleteMetricsDefsApi,
} = generateApis<
  EntityType.MetricsDefinition,
  Partial<MetricsDefinitionEntity>
>(
  [EntityType.MetricsDefinition, "metricsDefinition", "metrics"],
  [addManyMetricsDefs, addOneMetricsDef, updateMetricsDef, removeMetricsDef],
  [],
  [undefined, handleMetricsDefDelete]
);
export const createMetricsDefsAndFocusApi = createAsyncThunk(
  `${EntityType.MetricsDefinition}/fetchManyMetricsDefsAndFocusApi`,
  async (args: Partial<MetricsDefinitionEntity>, thunkAPI) => {
    const { payload: createdMetricsDef } = await thunkAPI.dispatch(
      createMetricsDefsApi(args)
    );
    await dataModelerService.dispatch("setActiveAsset", [
      EntityType.MetricsDefinition,
      createdMetricsDef.id,
    ]);
    return createdMetricsDef;
  }
);
export const updateMetricsDefsWrapperApi = invalidateExplorerThunk(
  EntityType.MetricsDefinition,
  updateMetricsDefsApi,
  ["sourceModelId", "timeDimension"],
  (state, id) => [id]
);

export const generateMeasuresAndDimensionsApi = createAsyncThunk(
  `${EntityType.MetricsDefinition}/generateMeasuresAndDimensions`,
  async (id: string, thunkAPI) => {
    const stream = streamingFetchWrapper<
      MeasureDefinitionEntity | DimensionDefinitionEntity
    >(`metrics/${id}/generate-measures-dimensions`, "POST");
    thunkAPI.dispatch(clearMeasuresForMetricsDefId(id));
    thunkAPI.dispatch(clearDimensionsForMetricsDefId(id));
    await asyncWait(10);
    for await (const measureOrDimension of stream) {
      if (measureOrDimension.type === EntityType.MeasureDefinition) {
        thunkAPI.dispatch(
          addOneMeasure(measureOrDimension as MeasureDefinitionEntity)
        );
      } else {
        thunkAPI.dispatch(
          addOneDimension(measureOrDimension as DimensionDefinitionEntity)
        );
      }
    }
  }
);

export const validateSelectedSources = createAsyncThunk(
  `${EntityType.MetricsDefinition}/validateSelectedSources`,
  async (id: string, thunkAPI) => {
    const state = thunkAPI.getState() as RillReduxState;
    const metricsDefinition = selectMetricsDefinitionById(state, id);

    let sourceModelValidationStatus: SourceModelValidationStatus;
    let derivedModel: DerivedModelEntity;
    if (metricsDefinition.sourceModelId) {
      // if some source model is selected, pull the derived model from the derived model state.
      derivedModel = selectDerivedModelById(metricsDefinition.sourceModelId);
      if (derivedModel) {
        // if a model is found, mark as INVALID if it has error
        sourceModelValidationStatus = derivedModel.error
          ? SourceModelValidationStatus.INVALID
          : SourceModelValidationStatus.OK;
      } else {
        // no model was found, most probably selected model was deleted
        sourceModelValidationStatus = SourceModelValidationStatus.MISSING;
      }
    } else {
      // empty selection will not throw error for now.
      sourceModelValidationStatus = SourceModelValidationStatus.OK;
    }
    thunkAPI.dispatch(
      setSourceModelValidationStatus(id, sourceModelValidationStatus)
    );

    let timeDimensionValidationStatus: SourceModelValidationStatus;
    if (metricsDefinition.timeDimension) {
      if (
        derivedModel &&
        derivedModel.profile?.find(
          (column) => column.name === metricsDefinition.timeDimension
        )
      ) {
        // if a model is found, mark as INVALID if it has error
        timeDimensionValidationStatus =
          !derivedModel.error &&
          derivedModel.profile?.find(
            (column) => column.name === metricsDefinition.timeDimension
          )
            ? SourceModelValidationStatus.OK
            : SourceModelValidationStatus.INVALID;
      } else {
        // no model was found, most probably selected model was deleted
        timeDimensionValidationStatus = SourceModelValidationStatus.MISSING;
      }
    } else {
      // empty selection will not throw error for now.
      timeDimensionValidationStatus = SourceModelValidationStatus.OK;
    }
    thunkAPI.dispatch(
      setTimeDimensionValidationStatus(id, timeDimensionValidationStatus)
    );

    // metrics explorer is active and model is no longer valid switch back to metrics definition
    if (
      state.application.activeEntity.id === id &&
      state.application.activeEntity.type === EntityType.MetricsExplorer &&
      sourceModelValidationStatus !== SourceModelValidationStatus.OK
    ) {
      await dataModelerService.dispatch("setActiveAsset", [
        EntityType.MetricsDefinition,
        id,
      ]);
    }

    // trigger measure and dimension validations
    selectMeasuresByMetricsId(state, id).forEach((measure) =>
      validateMeasureExpression(
        thunkAPI.dispatch,
        id,
        measure.id,
        selectMeasureById(state, measure.id).expression
      )
    );
    selectDimensionsByMetricsId(state, id).forEach((dimension) =>
      thunkAPI.dispatch(validateDimensionColumnApi(dimension.id))
    );
    // TODO: if timestamp column is invalid select the next valid timestamp column
  }
);
