import { goto } from "$app/navigation";
import type { DerivedModelEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import type { DimensionDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { MeasureDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { MetricsDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { SourceModelValidationStatus } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { asyncWait } from "@rilldata/web-local/common/utils/waitUtils";
import { validateDimensionColumnApi } from "../dimension-definition/dimension-definition-apis";
import { selectDimensionsByMetricsId } from "../dimension-definition/dimension-definition-selectors";
import {
  addOneDimension,
  clearDimensionsForMetricsDefId,
} from "../dimension-definition/dimension-definition-slice";
import { validateMeasureExpressionApi } from "../measure-definition/measure-definition-apis";
import {
  selectMeasureById,
  selectMeasuresByMetricsId,
} from "../measure-definition/measure-definition-selectors";
import {
  addOneMeasure,
  clearMeasuresForMetricsDefId,
} from "../measure-definition/measure-definition-slice";
import { selectMetricsDefinitionById } from "./metrics-definition-selectors";
import {
  addManyMetricsDefs,
  addOneMetricsDef,
  removeMetricsDef,
  setSourceModelValidationStatus,
  setTimeDimensionValidationStatus,
  updateMetricsDef,
} from "./metrics-definition-slice";
import { selectDerivedModelById } from "../model/model-selector";
import { createAsyncThunk } from "../redux-toolkit-wrapper";
import { selectTimestampColumnFromProfileEntity } from "../source/source-selectors";
import type { RillReduxState } from "../store-root";
import { generateApis } from "../utils/api-utils";
import { invalidateExplorer } from "../utils/invalidateExplorerThunk";
import { streamingFetchWrapper } from "../../util/fetchWrapper";

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
  [undefined, undefined]
);
export const createMetricsDefsAndFocusApi = createAsyncThunk(
  `${EntityType.MetricsDefinition}/fetchManyMetricsDefsAndFocusApi`,
  async (args: Partial<MetricsDefinitionEntity>, thunkAPI) => {
    const { payload: createdMetricsDef } = await thunkAPI.dispatch(
      createMetricsDefsApi(args)
    );
    goto(`/dashboard/${createdMetricsDef.id}/edit`);
    return createdMetricsDef;
  }
);
export const updateMetricsDefsWrapperApi = createAsyncThunk(
  `${EntityType.MetricsDefinition}/updateWrapperApi`,
  async (
    { id, changes }: { id: string; changes: Partial<MetricsDefinitionEntity> },
    thunkAPI
  ) => {
    if ("sourceModelId" in changes) {
      changes.timeDimension = selectTimestampColumnFromProfileEntity(
        selectDerivedModelById(changes.sourceModelId)
      )[0]?.name;
    }
    await invalidateExplorer(
      id,
      changes,
      thunkAPI,
      EntityType.MetricsDefinition,
      updateMetricsDefsApi,
      ["sourceModelId", "timeDimension"],
      (state, id) => [id]
    );
    if ("sourceModelId" in changes || "timeDimension" in changes) {
      thunkAPI.dispatch(validateSelectedSources(id));
    }
  }
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
      (sourceModelValidationStatus !== SourceModelValidationStatus.OK ||
        timeDimensionValidationStatus !== SourceModelValidationStatus.OK)
    ) {
      goto(`/dashboard/${id}/edit`);
    }

    // trigger measure and dimension validations
    await Promise.all(
      selectMeasuresByMetricsId(state, id).map((measure) =>
        thunkAPI.dispatch(
          validateMeasureExpressionApi({
            metricsDefId: id,
            measureId: measure.id,
            expression: selectMeasureById(state, measure.id).expression,
          })
        )
      )
    );
    await Promise.all(
      selectDimensionsByMetricsId(state, id).map((dimension) =>
        thunkAPI.dispatch(validateDimensionColumnApi(dimension.id))
      )
    );
    // TODO: if timestamp column is invalid select the next valid timestamp column
  }
);
