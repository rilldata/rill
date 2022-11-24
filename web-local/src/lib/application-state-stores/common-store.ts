import type { V1MigrationError } from "@rilldata/web-common/runtime-client";
import { Readable, writable } from "svelte/store";

export type CommonEntityData = {
  errors: Array<V1MigrationError>;
};

/**
 * Store to save common data across entities.
 * Currently only has errors
 */
export type CommonEntityState = {
  entities: Record<string, CommonEntityData>;
};
const { update, subscribe } = writable({
  entities: {},
} as CommonEntityState);

const createOrUpdateEntity = (
  path: string,
  callback: (entityData: CommonEntityData) => void
) => {
  update((state) => {
    if (!state[path]) {
      state.entities[path] = {
        errors: [],
      };
    }
    callback(state.entities[path]);
    return state;
  });
};

const commonEntitiesReducers = {
  setErrors(path: string, errors: Array<V1MigrationError>) {
    createOrUpdateEntity(path, (entityData: CommonEntityData) => {
      entityData.errors = errors;
    });
  },

  consolidateMigrateResponse(
    affectedPaths: Array<string>,
    errors: Array<V1MigrationError>
  ) {
    const errorsForPaths = new Map<string, Array<V1MigrationError>>();
    errors.forEach((error) => {
      if (!errorsForPaths.has(error.filePath)) {
        errorsForPaths.set(error.filePath, []);
      }
      errorsForPaths.get(error.filePath).push(error);
    });

    errorsForPaths.forEach((errors, path) => {
      commonEntitiesStore.setErrors(path, errors);
    });
  },
};

export const commonEntitiesStore: Readable<CommonEntityState> &
  typeof commonEntitiesReducers = {
  subscribe,
  ...commonEntitiesReducers,
};
