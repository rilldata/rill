import { useCanvasValidSpec } from "@rilldata/web-common/features/canvas/selector";
import {
  type RpcStatus,
  type V1CanvasSpec,
} from "@rilldata/web-common/runtime-client";
import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient, QueryObserverResult } from "@tanstack/svelte-query";
import { getContext } from "svelte";
import { derived, writable, type Readable, type Writable } from "svelte/store";
import type { CanvasEntity } from "../stores/canvas-entity";
import { useCanvasStore } from "../stores/canvas-stores";

export type StateManagers = {
  runtime: Writable<Runtime>;
  canvasName: Writable<string>;
  canvasStore: Readable<CanvasEntity>;
  validSpecStore: Readable<
    QueryObserverResult<V1CanvasSpec | undefined, RpcStatus>
  >;
  queryClient: QueryClient;
};

export const DEFAULT_STORE_KEY = Symbol("canvas-state-managers");

export function getCanvasStateManagers(): StateManagers {
  return getContext(DEFAULT_STORE_KEY);
}

export function createStateManagers({
  queryClient,
  canvasName,
}: {
  queryClient: QueryClient;
  canvasName: string;
}): StateManagers {
  const canvasNameStore = writable(canvasName);

  const canvasStore: Readable<CanvasEntity> = derived(
    [canvasNameStore],
    ([name], set) => {
      const store = useCanvasStore(name);
      return store.subscribe(set);
    },
  );

  const validSpecStore: Readable<
    QueryObserverResult<V1CanvasSpec | undefined, RpcStatus>
  > = derived([runtime, canvasNameStore], ([r, canvasName], set) =>
    useCanvasValidSpec(r.instanceId, canvasName, { queryClient }).subscribe(
      set,
    ),
  );

  return {
    runtime: runtime,
    canvasName: canvasNameStore,
    canvasStore,
    validSpecStore,
    queryClient,
  };
}
