import { EntityStateUpdatesHandler } from "./EntityStateUpdatesHandler";
import type { PersistentModelEntity } from "../entity-state-service/PersistentModelEntityService";

export class PersistentModelUpdateHandler extends EntityStateUpdatesHandler<PersistentModelEntity> {
  public async handleUpdatedEntity(
    modelEntity: PersistentModelEntity
  ): Promise<void> {
    await this.dataModelerService.dispatch("updateModelQuery", [
      modelEntity.id,
      modelEntity.query,
    ]);
  }
}
