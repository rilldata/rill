import type { PersistentTableEntityService } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import type {
  ApplicationState,
  ApplicationStateActionArg,
} from "../data-modeler-state-service/entity-state-service/ApplicationEntityService";
import {
  EntityRecord,
  EntityStateActionArg,
  EntityStateService,
  EntityType,
  StateType,
} from "../data-modeler-state-service/entity-state-service/EntityStateService";
import type { PersistentModelEntityService } from "../data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type { PersistentModelStateActionArg } from "../data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import {
  DatabaseActionQueuePriority,
  DatabaseProfilesFieldPriority,
  getProfilePriority,
  MetadataPriority,
  ProfileMetadataPriorityMap,
} from "../priority-action-queue/DatabaseActionQueuePriority";
import { DataModelerActions } from "./DataModelerActions";

export class ApplicationActions extends DataModelerActions {
  @DataModelerActions.ApplicationAction()
  public async setActiveAsset(
    { stateService }: ApplicationStateActionArg,
    entityType: EntityType,
    entityId: string
  ) {
    const currentActiveAsset = (
      stateService.getCurrentState() as ApplicationState
    ).activeEntity;

    // mark older model as inactive.
    if (
      currentActiveAsset?.type === EntityType.Model &&
      currentActiveAsset?.id
    ) {
      let columns;
      try {
        columns = this.getEntityColumns(
          EntityType.Model,
          currentActiveAsset.id
        );

        columns.forEach((column) => {
          Object.values(MetadataPriority).forEach((priority) => {
            this.databaseActionQueue.updatePriority(
              currentActiveAsset.id + column + priority,
              getProfilePriority(
                DatabaseActionQueuePriority.InactiveModelProfile,
                DatabaseProfilesFieldPriority.NonFocused,
                ProfileMetadataPriorityMap[priority]
              )
            );
          });
        });
      } catch (e) {
        // swallow error for now
      }
    }

    // upgrade profile priority of newly selected asset
    if (entityType === EntityType.Model) {
      try {
        const columns = this.getEntityColumns(EntityType.Model, entityId);

        columns?.forEach((column) => {
          Object.values(MetadataPriority).forEach((priority) => {
            this.databaseActionQueue.updatePriority(
              currentActiveAsset.id + column + priority,
              getProfilePriority(
                DatabaseActionQueuePriority.ActiveModelProfile,
                DatabaseProfilesFieldPriority.NonFocused,
                ProfileMetadataPriorityMap[priority]
              )
            );
          });
        });
      } catch (e) {
        // swallow error for now
      }
    }

    this.dataModelerStateService.dispatch("setActiveAsset", [
      entityType,
      entityId,
    ]);
  }

  @DataModelerActions.ApplicationAction()
  public async updateFocusProfilePriority(
    _: ApplicationStateActionArg,
    entityId: string,
    column: string
  ) {
    Object.values(MetadataPriority).forEach((priority) => {
      this.databaseActionQueue.updatePriority(
        entityId + column + priority,
        getProfilePriority(
          DatabaseActionQueuePriority.ActiveModelProfile,
          DatabaseProfilesFieldPriority.Focused,
          ProfileMetadataPriorityMap[priority]
        )
      );
    });
  }

  @DataModelerActions.ApplicationAction()
  public async clearColumnProfilePriority(
    _: ApplicationStateActionArg,
    entityType: EntityType,
    entityId: string
  ) {
    const columns = this.getEntityColumns(entityType, entityId);
    columns.forEach((column) => {
      Object.values(MetadataPriority).forEach((priority) => {
        this.databaseActionQueue.clearQueue(entityId + column + priority);
      });
    });
  }

  @DataModelerActions.PersistentModelAction()
  public async setModelAsActiveAsset({
    stateService,
  }: PersistentModelStateActionArg) {
    this.dataModelerStateService.dispatch("setActiveAsset", [
      EntityType.Model,
      stateService.getCurrentState().entities[0]?.id,
    ]);
  }

  @DataModelerActions.PersistentAction()
  public async deleteEntity(
    { stateService }: EntityStateActionArg<EntityRecord>,
    entityType: EntityType,
    entityName: string
  ) {
    const entity = this.getEntityByName(stateService, entityName);
    if (!entity) {
      return;
    }

    this.databaseActionQueue.clearQueue(entityName);
    this.dataModelerService.dispatch("clearColumnProfilePriority", [
      entityType,
      entity.id,
    ]);

    this.dataModelerStateService.dispatch("deleteEntity", [
      entityType,
      StateType.Persistent,
      entity.id,
    ]);
    this.dataModelerStateService.dispatch("deleteEntity", [
      entityType,
      StateType.Derived,
      entity.id,
    ]);
  }

  // Temporary until nodejs is removed
  @DataModelerActions.PersistentAction()
  public async renameEntity(
    { stateService }: EntityStateActionArg<EntityRecord>,
    entityType: EntityType,
    fromName: string,
    toName: string
  ) {
    const entity = this.getEntityByName(stateService, fromName);
    if (!entity) {
      return;
    }

    switch (entityType) {
      case EntityType.Model:
        this.dataModelerStateService.dispatch("updateModelName", [
          entity.id,
          toName,
        ]);
        break;
      case EntityType.Table:
        this.dataModelerStateService.dispatch("updateTableName", [
          entity.id,
          toName,
        ]);
        break;
    }

    this.databaseActionQueue.clearQueue(fromName);
  }

  private getEntityColumns(entityType: EntityType, entityId: string) {
    if (entityType === EntityType.Table || entityType === EntityType.Model) {
      return (
        this.dataModelerStateService
          .getEntityStateService(entityType, StateType.Derived)
          .getById(entityId)
          .profile?.map((column) => column.name) || []
      );
    } else return [];
  }

  private getEntityByName(
    stateService: EntityStateService<EntityRecord>,
    entityName: string
  ) {
    let entity: EntityRecord;
    switch (stateService.entityType) {
      case EntityType.Model:
        entity = (stateService as PersistentModelEntityService).getByField(
          "tableName",
          entityName
        );
        break;
      case EntityType.Table:
        entity = (stateService as PersistentTableEntityService).getByField(
          "tableName",
          entityName
        );
        break;
    }
    return entity;
  }
}
