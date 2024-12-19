import {
  type V1Canvas,
  type V1CanvasSpec,
} from "@rilldata/web-common/runtime-client";
import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { getContext } from "svelte";
import {
  derived,
  get,
  readable,
  writable,
  type Readable,
  type Writable,
} from "svelte/store";
import type { CanvasEntity } from "../stores/canvas-entity";
import { updateCanvasByName, useCanvasStore } from "../stores/canvas-stores";
import type { CanvasCallbackExecutor } from "./types";

export type StateManagers = {
  runtime: Writable<Runtime>;
  canvasName: Writable<string>;
  canvasStore: Readable<CanvasEntity>;
  validSpecStore: Readable<V1CanvasSpec>;
  updateCanvas: CanvasCallbackExecutor;
  queryClient: QueryClient;
};

export const DEFAULT_STORE_KEY = Symbol("canvas-state-managers");

export function getCanvasStateManagers(): StateManagers {
  return getContext(DEFAULT_STORE_KEY);
}

export function createStateManagers({
  queryClient,
  canvasName,
  canvasResource,
}: {
  queryClient: QueryClient;
  canvasName: string;
  canvasResource: V1Canvas;
}): StateManagers {
  const canvasNameStore = writable(canvasName);

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
    canvasStore,
    validSpecStore,
    queryClient,
    updateCanvas,
  };
}
