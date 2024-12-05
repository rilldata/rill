import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import {
  type V1Canvas,
  type V1CanvasSpec,
} from "@rilldata/web-common/runtime-client";
import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { getContext } from "svelte";
import {
  type Readable,
  type Writable,
  derived,
  get,
  readable,
  writable,
} from "svelte/store";
import type { CanvasEntity } from "../stores/canvas-entity";
import { updateCanvasByName, useCanvasStore } from "../stores/canvas-stores";
import type { CanvasCallbackExecutor } from "./types";

export type StateManagers = {
  runtime: Writable<Runtime>;
  canvasName: Writable<string>;
  canvasStore: Readable<CanvasEntity>;
  fileArtifact: Readable<FileArtifact>;
  validSpecStore: Readable<V1CanvasSpec>;
  updateCanvas: CanvasCallbackExecutor;
  queryClient: QueryClient;

  // Enable later
  /**
   * A collection of Readables that can be used to select data from the dashboard.
   */
  // selectors: StateManagerReadables;
  /**
   * A collection of functions that update the dashboard data model.
   */
  // actions: StateManagerActions;
};

export const DEFAULT_STORE_KEY = Symbol("canvas-state-managers");

export function getCanvasStateManagers(): StateManagers {
  return getContext(DEFAULT_STORE_KEY);
}

export function createStateManagers({
  queryClient,
  canvasName,
  canvasResource,
  fileArtifact,
}: {
  queryClient: QueryClient;
  canvasName: string;
  canvasResource: V1Canvas;
  fileArtifact: FileArtifact;
}): StateManagers {
  console.log("fileArtifact", fileArtifact);
  const canvasNameStore = writable(canvasName);

  const fileArtifactStore = readable(fileArtifact);
  const canvasStore: Readable<CanvasEntity> = derived(
    [canvasNameStore],
    ([name], set) => {
      const store = useCanvasStore(name);
      return store.subscribe(set);
    },
  );

  const validSpecStore = readable(canvasResource?.state?.validSpec);

  const updateCanvas = (callback: (canvasEntity: CanvasEntity) => void) => {
    const name = get(canvasStore).name;
    updateCanvasByName(name, callback);
  };

  return {
    runtime: runtime,
    canvasName: canvasNameStore,
    fileArtifact: fileArtifactStore,
    canvasStore,
    validSpecStore,
    queryClient,
    updateCanvas,
  };
}
