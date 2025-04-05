import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { type Writable } from "svelte/store";
import { useCanvasEntity } from "../stores/canvas-entities";
import type { CanvasEntity } from "../stores/canvas-entity";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export type CanvasStore = {
  runtime: Writable<Runtime>;
  canvasEntity: CanvasEntity;
  queryClient: QueryClient;
};

const canvasStores = new Map<string, CanvasStore>();

export function getCanvasStore(canvasName: string) {
  let store = canvasStores.get(canvasName);
  if (!store) {
    store = createCanvasStore(canvasName);
    canvasStores.set(canvasName, store);
  }
  return store;
}

export function createCanvasStore(canvasName: string): CanvasStore {
  const canvasEntity = useCanvasEntity(canvasName);

  return {
    runtime: runtime,
    canvasEntity,
    queryClient,
  };
}
