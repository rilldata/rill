import type { V1ReconcileError } from "@rilldata/web-common/runtime-client";
import { Readable, writable } from "svelte/store";

export type FileArtifactsData = {
  errors: Array<V1ReconcileError>;
};

/**
 * Store to save data for each file artifact.
 * Currently, stores reconcile errors
 */
export type FileArtifactsState = {
  entities: Record<string, FileArtifactsData>;
};
const { update, subscribe } = writable({
  entities: {},
} as FileArtifactsState);

const createOrUpdateFileArtifact = (
  path: string,
  callback: (entityData: FileArtifactsData) => void
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

const fileArtifactsEntitiesReducers = {
  setErrors(affectedPaths: Array<string>, errors: Array<V1ReconcileError>) {
    const errorsForPaths = new Map<string, Array<V1ReconcileError>>();
    affectedPaths.forEach((affectedPath) =>
      errorsForPaths.set(correctFilePath(affectedPath), [])
    );

    errors.forEach((error) => {
      const filePath = correctFilePath(error.filePath);

      if (!errorsForPaths.has(filePath)) {
        errorsForPaths.set(filePath, []);
      }
      errorsForPaths.get(filePath).push(error);
    });

    errorsForPaths.forEach((errors, path) => {
      createOrUpdateFileArtifact(path, (entityData: FileArtifactsData) => {
        entityData.errors = errors;
      });
    });
  },
};

export const fileArtifactsStore: Readable<FileArtifactsState> &
  typeof fileArtifactsEntitiesReducers = {
  subscribe,
  ...fileArtifactsEntitiesReducers,
};

function correctFilePath(filePath: string) {
  // TODO: why does affectedPaths not have the leading /
  if (!filePath.startsWith("/")) {
    filePath = "/" + filePath;
  }
  return filePath;
}
