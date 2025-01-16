import { CanvasEntity } from "./canvas-entity";

export class CanvasEntities {
  private entities: Map<string, CanvasEntity> = new Map();

  addCanvas(name: string) {
    if (!this.entities.has(name)) {
      this.entities.set(name, new CanvasEntity(name));
    }
  }

  removeCanvas(name: string) {
    this.entities.delete(name);
  }

  hasCanvas(name: string) {
    return this.entities.has(name);
  }

  getCanvas(name: string) {
    let canvasEntity = this.entities.get(name);

    if (!canvasEntity) {
      canvasEntity = new CanvasEntity(name);
      this.entities.set(name, canvasEntity);
    }

    return canvasEntity;
  }
}

export const canvasEntities = new CanvasEntities();

export function useCanvasEntity(name: string) {
  return canvasEntities.getCanvas(name);
}
