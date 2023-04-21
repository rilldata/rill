import type { V1ReconcileError } from "@rilldata/web-common/runtime-client";
import { derived, Readable, writable } from "svelte/store";
import { parseDocument } from "yaml";
import type { MetricsConfig } from "../metrics-views/metrics-internal-store";

export type FileArtifactsData = {
  // name used for the file. will not always be dependent on the file
  name?: string;
  errors?: Array<V1ReconcileError>;
  jsonRepresentation?: MetricsConfig | Record<string, never>;
  isReconciling?: boolean;
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
    if (!state.entities[path]) {
      state.entities[path] = {
        errors: [],
        jsonRepresentation: {},
        isReconciling: false,
      };
    }
    callback(state.entities[path]);
    return state;
  });
};

const fileArtifactsEntitiesReducers = {
  setName(path: string, name: string) {
    createOrUpdateFileArtifact(path, (entityData) => {
      entityData.name = name;
    });
  },

  // setRuntimeError(path: string, errors: Array<V1ReconcileError>) {
  //   createOrUpdateFileArtifact(path, (entityData) => {
  //     entityData.errors = errors;
  //   });
  // },

  // SK: Only handles embedded source errors, not model entity errors (those errors don't have file paths)
  setErrors(affectedPaths: Array<string>, errors: Array<V1ReconcileError>) {
    const errorsForPaths = new Map<string, Array<V1ReconcileError>>();
    affectedPaths.forEach((affectedPath) =>
      errorsForPaths.set(correctFilePath(affectedPath), [])
    );

    errors.forEach((error) => {
      const filePath = correctFilePath(error.filePath);

      // empty models should not show error
      if (
        filePath.endsWith(".sql") &&
        error.message.endsWith("No statement to prepare!")
      ) {
        return;
      }

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

  // reducer for storing Metric artifact
  setJSONRep(affectedPath: string, fileData: string) {
    const jsonRepresentation = parseDocument(fileData).toJSON();
    createOrUpdateFileArtifact(
      affectedPath,
      (entityData: FileArtifactsData) => {
        entityData.jsonRepresentation = jsonRepresentation;
      }
    );
  },

  setIsReconciling(affectedPath: string, isReconciling: boolean) {
    createOrUpdateFileArtifact(
      affectedPath,
      (entityData: FileArtifactsData) => {
        entityData.isReconciling = isReconciling;
      }
    );
  },
};

export type FileArtifactsStore = Readable<FileArtifactsState> &
  typeof fileArtifactsEntitiesReducers;

export const fileArtifactsStore: FileArtifactsStore = {
  subscribe,
  ...fileArtifactsEntitiesReducers,
};

export function getIsFileReconcilingStore(file: string) {
  return derived(
    fileArtifactsStore,
    ($store) => $store.entities[file]?.isReconciling
  );
}

function correctFilePath(filePath: string) {
  // TODO: why does affectedPaths not have the leading /
  if (!filePath.startsWith("/")) {
    filePath = "/" + filePath;
  }
  return filePath;
}
