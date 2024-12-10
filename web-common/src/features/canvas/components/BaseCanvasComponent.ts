import type {
  CanvasComponent,
  ComponentSize,
} from "@rilldata/web-common/features/canvas/components/component-types";
import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import { writable, type Writable } from "svelte/store";

// A base class that implements all the store logic
export abstract class BaseCanvasComponent<T> implements CanvasComponent<T> {
  specStore: Writable<T>;

  // Let child classes define these
  abstract minSize: ComponentSize;
  abstract defaultSize: ComponentSize;
  abstract isValid(spec: T): boolean;
  abstract inputParams(): Record<keyof T, ComponentInputParam>;

  constructor(defaultSpec: T, initialSpec: Partial<T> = {}) {
    // Initialize the store with merged spec
    const mergedSpec = { ...defaultSpec, ...initialSpec };
    this.specStore = writable(mergedSpec);
  }

  setSpec(newSpec: T): void {
    this.specStore.set(newSpec);
  }

  updateSpec(updater: (spec: T) => T): void {
    this.specStore.update(updater);
  }
}
