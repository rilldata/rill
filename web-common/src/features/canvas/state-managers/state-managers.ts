import { useCanvasValidSpec } from "@rilldata/web-common/features/canvas/selector";
import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { getContext } from "svelte";
import { derived, writable, type Writable } from "svelte/store";
import { useCanvasEntity } from "../stores/canvas-entities";
import type { CanvasEntity } from "../stores/canvas-entity";

export type StateManagers = {
  runtime: Writable<Runtime>;
  canvasName: Writable<string>;
  canvasEntity: CanvasEntity;
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

  const validSpecStore: CanvasSpecResponseStore = derived(
    [runtime, canvasNameStore],
    ([r, canvasName], set) =>
      useCanvasValidSpec(r.instanceId, canvasName, { queryClient }).subscribe(
        set,
      ),
  );

  const canvasEntity = useCanvasEntity(canvasName, validSpecStore);

  return {
    runtime: runtime,
    canvasName: canvasNameStore,
    canvasEntity,
    queryClient,
  };
}
