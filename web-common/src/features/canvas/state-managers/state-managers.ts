import type { QueryClient } from "@tanstack/svelte-query";
import { CanvasEntity } from "../stores/canvas-entity";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { CanvasResponse } from "../selector";

export type CanvasStore = {
  canvasEntity: CanvasEntity;
  queryClient: QueryClient;
};

type CanvasId = `${string}::${string}`;

const canvasRegistry = new Map<string, CanvasStore>();

function makeCanvasId(canvasName: string, instanceId: string): CanvasId {
  return `${instanceId}::${canvasName}`;
}

export function getCanvasStoreUnguarded(
  canvasName: string,
  instanceId: string,
): CanvasStore | undefined {
  const id = makeCanvasId(canvasName, instanceId);

  return canvasRegistry.get(id);
}

export function getCanvasStore(
  canvasName: string,
  instanceId: string,
): CanvasStore {
  const id = makeCanvasId(canvasName, instanceId);

  const store = canvasRegistry.get(id);

  if (!store) {
    throw new Error(
      `Canvas store for ID ${id} has not been initialized. Use 'setCanvasStore' first.`,
    );
  }

  return store;
}

export function removeCanvasStore(
  canvasName: string,
  instanceId: string,
): void {
  const id = makeCanvasId(canvasName, instanceId);
  canvasRegistry.delete(id);
}

export function setCanvasStore(
  canvasName: string,
  instanceId: string,
  response: CanvasResponse,
): CanvasStore {
  const id = makeCanvasId(canvasName, instanceId);

  if (canvasRegistry.has(id)) {
    console.warn(
      `Canvas store for ID ${id} already exists. Returning existing store.`,
    );
    return canvasRegistry.get(id)!;
  }

  const canvasEntity = new CanvasEntity(canvasName, instanceId, response);
  const store: CanvasStore = {
    canvasEntity,
    queryClient,
  };

  canvasRegistry.set(id, store);
  return store;
}
