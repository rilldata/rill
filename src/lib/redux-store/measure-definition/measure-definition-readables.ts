import { createReadableFactoryWithSelector } from "$lib/redux-store/svelte-readables-wrapper";
import { store } from "$lib/redux-store/store-root";
import {
  selectMeasureById,
  selectMeasureFieldNameByIdAndIndex,
  selectMeasuresByIds,
  selectMeasuresByMetricsId,
  selectValidMeasuresByMetricsId,
} from "$lib/redux-store/measure-definition/measure-definition-selectors";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";

export const getMeasureById = createReadableFactoryWithSelector<
  MeasureDefinitionEntity,
  [string]
>(store, selectMeasureById);

export const getMeasuresByMetricsId = createReadableFactoryWithSelector(
  store,
  selectMeasuresByMetricsId
);
export const getValidMeasuresByMetricsId = createReadableFactoryWithSelector(
  store,
  selectValidMeasuresByMetricsId
);

export const getMeasuresByIds = createReadableFactoryWithSelector(
  store,
  selectMeasuresByIds
);

export const getMeasureFieldNameByIdAndIndex =
  createReadableFactoryWithSelector(store, selectMeasureFieldNameByIdAndIndex);
