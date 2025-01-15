import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import { CanvasEntity } from "./canvas-entity";

export class CanvasEntities {
  private entities: Map<string, CanvasEntity> = new Map();

  addCanvas(name: string, validSpecStore: CanvasSpecResponseStore) {
    if (!this.entities.has(name)) {
      this.entities.set(name, new CanvasEntity(name, validSpecStore));
    }
  }

  removeCanvas(name: string) {
    this.entities.delete(name);
  }

  hasCanvas(name: string) {
    return this.entities.has(name);
  }

  getCanvas(name: string, validSpecStore: CanvasSpecResponseStore) {
    let canvasEntity = this.entities.get(name);

    if (!canvasEntity) {
      canvasEntity = new CanvasEntity(name, validSpecStore);
      this.entities.set(name, canvasEntity);
    }

    return canvasEntity;
  }
}

export const canvasEntities = new CanvasEntities();

export function useCanvasEntity(
  name: string,
  validSpecStore: CanvasSpecResponseStore,
) {
  return canvasEntities.getCanvas(name, validSpecStore);
}
