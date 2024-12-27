// CanvasStore.ts
import { derived, writable, type Readable } from "svelte/store";
import { CanvasEntity } from "./canvas-entity";

export interface CanvasStoreType {
  entities: Record<string, CanvasEntity>;
}

function createCanvasStore() {
  const { subscribe, update } = writable<CanvasStoreType>({
    entities: {},
  });

  // Add a new CanvasEntity to the store by name
  function addEntity(name: string) {
    update((store) => {
      // Only add if it doesn’t exist yet
      if (!store.entities[name]) {
        store.entities[name] = new CanvasEntity(name);
      }
      return store;
    });
  }

  // Remove an existing CanvasEntity by name
  function removeEntity(name: string) {
    update((store) => {
      delete store.entities[name];
      return store;
    });
  }

  return {
    subscribe,
    addEntity,
    removeEntity,
  };
}

// Export a singleton instance for convenience
export const canvasEntityStore = createCanvasStore();

export function useCanvasStore(name: string): Readable<CanvasEntity> {
  return derived(canvasEntityStore, ($canvasStore) => {
    return $canvasStore.entities[name];
  });
}
