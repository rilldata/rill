import type {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { EntityRecordMapType } from "$common/data-modeler-state-service/entity-state-service/EntityStateServicesMap";
import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { queryClient } from "$lib/svelte-query/globalQueryClient";
import { invalidateMetricsView } from "$lib/svelte-query/queries/metrics-view";
import type { AsyncThunk } from "@reduxjs/toolkit";

/**
 * Async thunk version that invalidates explorer.
 * Calls {@link invalidateExplorer} from within the thunk.
 */
export const invalidateExplorerThunk = <
  Type extends EntityType,
  Entity = EntityRecordMapType[Type][StateType.Persistent]
>(
  type: EntityType,
  updateApi: AsyncThunk<
    void,
    { id: string; changes: Partial<Entity> },
    Record<string, never>
  >,
  keys: Array<keyof Entity>,
  // this returns an array to support sources/models in the future.
  // single source/model can map to multiple explorers
  metricsExplorerIdsGetter: (state: RillReduxState, id: string) => Array<string>
) => {
  return createAsyncThunk(
    `${type}/updateWrapperApi`,
    async (
      { id, changes }: { id: string; changes: Partial<Entity> },
      thunkAPI
    ) => {
      await invalidateExplorer(
        id,
        changes,
        thunkAPI,
        type,
        updateApi,
        keys,
        metricsExplorerIdsGetter
      );
    }
  );
};

/**
 * Non async thunk version that invalidates explorer
 */
export const invalidateExplorer = async <
  Type extends EntityType,
  Entity = EntityRecordMapType[Type][StateType.Persistent]
>(
  id: string,
  changes: Partial<Entity>,
  thunkAPI,
  type: EntityType,
  updateApi: AsyncThunk<
    void,
    { id: string; changes: Partial<Entity> },
    Record<string, never>
  >,
  keys: Array<keyof Entity>,
  metricsExplorerIdsGetter: (state: RillReduxState, id: string) => Array<string>
) => {
  await thunkAPI.dispatch(updateApi({ id, changes }));
  if (keys.some((key) => key in changes)) {
    const metricsExplorerIds = metricsExplorerIdsGetter(
      thunkAPI.getState() as RillReduxState,
      id
    );
    metricsExplorerIds.forEach((metricsExplorerId) =>
      invalidateMetricsView(queryClient, metricsExplorerId)
    );
  }
};
