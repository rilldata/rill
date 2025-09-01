import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import type { Writable } from "svelte/store";
import { CanvasEntity } from "../stores/canvas-entity";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export type CanvasStore = {
  runtime: Writable<Runtime>;
  canvasEntity: CanvasEntity;
  queryClient: QueryClient;
};

const canvasStores = new Map<string, CanvasStore>();

type CanvasId = `${string}::${string}`;

function makeCanvasId(canvasName: string, instanceId: string): CanvasId {
  return `${instanceId}::${canvasName}`;
}

export function getCanvasStore(
  canvasName: string,
  instanceId: string,
): CanvasStore {
  const id = makeCanvasId(canvasName, instanceId);

  let store = canvasStores.get(id);

  if (!store) {
    store = createCanvasStore(canvasName, instanceId);
    canvasStores.set(id, store);
  }
  return store;
}

export function removeCanvasStore(
  canvasName: string,
  instanceId: string,
): void {
  const id = makeCanvasId(canvasName, instanceId);
  canvasStores.delete(id);
}

export function createCanvasStore(
  canvasName: string,
  instanceId: string,
): CanvasStore {
  const canvasEntity = new CanvasEntity(canvasName, instanceId);

  return {
    runtime: runtime,
    canvasEntity,
    queryClient,
  };
}
