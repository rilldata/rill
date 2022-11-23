import { StateActions } from "./StateActions";
import { DataModelerActions } from "../data-modeler-service/DataModelerActions";
import type { ApplicationStateActionArg } from "./entity-state-service/ApplicationEntityService";
import type { EntityType } from "./entity-state-service/EntityStateService";
import type { ApplicationStatus } from "./entity-state-service/ApplicationEntityService";

export class ApplicationStateActions extends StateActions {
  @DataModelerActions.ApplicationAction()
  public async setActiveAsset(
    { draftState }: ApplicationStateActionArg,
    entityType: EntityType,
    entityName: string
  ) {
    draftState.activeEntity = {
      type: entityType,
      id: entityName,
      name: entityName,
    };
  }

  @DataModelerActions.ApplicationAction()
  public async setApplicationStatus(
    { draftState }: ApplicationStateActionArg,
    status: ApplicationStatus
  ) {
    draftState.status = status;
  }

  @DataModelerActions.ApplicationAction()
  public async setDuckDbPath(
    { draftState }: ApplicationStateActionArg,
    duckDbPath: string
  ) {
    draftState.duckDbPath = duckDbPath;
  }
}
