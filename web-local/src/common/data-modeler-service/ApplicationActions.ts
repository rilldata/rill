import type {
  ApplicationState,
  ApplicationStateActionArg,
} from "../data-modeler-state-service/entity-state-service/ApplicationEntityService";
import {
  EntityRecord,
  EntityStateActionArg,
  EntityType,
  StateType,
} from "../data-modeler-state-service/entity-state-service/EntityStateService";
import type { PersistentModelStateActionArg } from "../data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import {
  DatabaseActionQueuePriority,
  DatabaseProfilesFieldPriority,
  getProfilePriority,
  MetadataPriority,
  ProfileMetadataPriorityMap,
} from "../priority-action-queue/DatabaseActionQueuePriority";
import { getNextEntityId } from "../utils/getNextEntityId";
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
      const columns = this.getEntityColumns(EntityType.Model, entityId);

      columns.forEach((column) => {
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
    entityId: string
  ) {
    const applicationState = this.dataModelerStateService.getApplicationState();
    if (
      applicationState.activeEntity?.id === entityId &&
      applicationState.activeEntity?.type === entityType
    ) {
      const newEntityId = getNextEntityId(
        stateService.getCurrentState().entities,
        entityId
      );
      if (newEntityId) {
        await this.dataModelerService.dispatch("setActiveAsset", [
          entityType,
          newEntityId,
        ]);
      }
    }

    this.databaseActionQueue.clearQueue(entityId);
    this.dataModelerService.dispatch("clearColumnProfilePriority", [
      entityType,
      entityId,
    ]);

    this.dataModelerStateService.dispatch("deleteEntity", [
      entityType,
      StateType.Persistent,
      entityId,
    ]);
    this.dataModelerStateService.dispatch("deleteEntity", [
      entityType,
      StateType.Derived,
      entityId,
    ]);
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
}
