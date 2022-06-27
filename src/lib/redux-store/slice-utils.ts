import * as reduxToolkit from "@reduxjs/toolkit";
import type { ActionCreatorWithPreparedPayload } from "@reduxjs/toolkit";
import type {
  EntityRecord,
  EntityType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { fetchWrapper } from "$lib/util/fetchWrapper";
import type { EntityRecordMapType } from "$common/data-modeler-state-service/entity-state-service/EntityStateServicesMap";
import type { StateType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { retry } from "@reduxjs/toolkit/query";

const { createAsyncThunk } = reduxToolkit;

function getQueryArgs(args: Record<string, any>) {
  if (!args) return "";
  return "/?" + Object.keys(args).map((argKey) => `${argKey}=${args[argKey]}`);
}

export function generateApis<
  Type extends EntityType,
  FetchManyParams extends Record<string, any> = Record<string, unknown>,
  CreateParams extends Record<string, any> = Record<string, unknown>,
  Entity = EntityRecordMapType[Type][StateType.Persistent]
>(
  entityType: EntityType,
  addManyAction: ActionCreatorWithPreparedPayload<
    [entities: Array<Entity>],
    Array<Entity>
  >,
  addOneAction: ActionCreatorWithPreparedPayload<[entity: Entity], Entity>,
  updateAction: ActionCreatorWithPreparedPayload<
    [id: string, changes: Partial<Entity>],
    { id: string; changes: Partial<Entity> }
  >,
  removeAction: ActionCreatorWithPreparedPayload<[id: string], string>,
  endpoint: string
) {
  return {
    fetchManyApi: createAsyncThunk(
      `${entityType}/fetchManyApi`,
      async (args: FetchManyParams, thunkAPI) => {
        thunkAPI.dispatch(
          addManyAction(
            await fetchWrapper(`${endpoint}${getQueryArgs(args)}`, "GET")
          )
        );
      },
      {}
    ),
    createApi: createAsyncThunk(
      `${entityType}/createApi`,
      async (args: CreateParams, thunkAPI) => {
        thunkAPI.dispatch(
          addOneAction(await fetchWrapper(endpoint, "PUT", args))
        );
      }
    ),
    updateApi: createAsyncThunk(
      `${entityType}/updateApi`,
      async (
        { id, changes }: { id: string; changes: Partial<Entity> },
        thunkAPI
      ) => {
        thunkAPI.dispatch(
          updateAction(
            id,
            await fetchWrapper(`${endpoint}/${id}`, "POST", changes)
          )
        );
      }
    ),
    deleteApi: createAsyncThunk(
      `${entityType}/deleteApi`,
      async (id: string, thunkAPI) => {
        await fetchWrapper(`${endpoint}/${id}`, "DELETE");
        thunkAPI.dispatch(removeAction(id));
      }
    ),
  };
}

export function generateBasicSelectors(sliceKey: keyof RillReduxState) {
  return {
    manySelector: (state: RillReduxState) =>
      state[sliceKey].ids.map((id) => state[sliceKey].entities[id]),
    singleSelector: (id: string) => {
      return (state: RillReduxState) => state[sliceKey].entities[id];
    },
  };
}

export function generateFilteredSelectors<FilterArgs extends Array<unknown>>(
  sliceKey: keyof RillReduxState,
  filter: (entity: unknown, ...args: FilterArgs) => boolean
) {
  return {
    manySelector: (...args: FilterArgs) => {
      return (state: RillReduxState) =>
        state[sliceKey].ids
          .filter((id) => filter(state[sliceKey].entities[id], ...args))
          .map((id) => state[sliceKey].entities[id]);
    },
    singleSelector: (id: string) => {
      return (state: RillReduxState) => state[sliceKey].entities[id];
    },
  };
}
