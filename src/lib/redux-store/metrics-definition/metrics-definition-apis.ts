import type {
  DerivedModelEntity,
  DerivedModelState,
} from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
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

const handleMetricsDefCreate = async (
  createdMetricsDef: MetricsDefinitionEntity
) => {
  await dataModelerService.dispatch("setActiveAsset", [
    EntityType.MetricsDefinition,
    createdMetricsDef.id,
  ]);
};
const handleMetricsDefDelete = async (id: string) => {
  const activeEntity = selectApplicationActiveEntity(store.getState());
  if (!activeEntity) return;

  if (
    activeEntity.id === id &&
    activeEntity.type === EntityType.MetricsDefinition
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
  [handleMetricsDefCreate, handleMetricsDefDelete]
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
  async (
    {
      id,
      derivedModelState,
    }: { id: string; derivedModelState: DerivedModelState },
    thunkAPI
  ) => {
    const metricsDefinition = selectMetricsDefinitionById(
      thunkAPI.getState() as RillReduxState,
      id
    );

    let sourceModelValidationStatus: SourceModelValidationStatus;
    let derivedModel: DerivedModelEntity;
    if (metricsDefinition.sourceModelId) {
      // if some source model is selected, pull the derived model from the derived model state.
      derivedModel = selectDerivedModelById(
        derivedModelState,
        metricsDefinition.sourceModelId
      );
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
  }
);
