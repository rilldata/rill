import type { DataModelerStateService } from "../data-modeler-state-service/DataModelerStateService";
import type { DataModelerService } from "./DataModelerService";
import type { DatabaseService } from "../database-service/DatabaseService";
import type { NotificationService } from "../notifications/NotificationService";
import { ActionsBase } from "../ActionsBase";
import type { RootConfig } from "../config/RootConfig";
import type { ActionQueueOrchestrator } from "../priority-action-queue/ActionQueueOrchestrator";
import type { DatabaseActionsDefinition } from "../database-service/DatabaseService";

/**
 * Class that has the actual action implementations.
 */
export class DataModelerActions extends ActionsBase {
  protected dataModelerService: DataModelerService;
  protected notificationService: NotificationService;
  protected databaseActionQueue: ActionQueueOrchestrator<DatabaseActionsDefinition>;

  constructor(
    protected readonly config: RootConfig,
    protected readonly dataModelerStateService: DataModelerStateService,
    protected readonly databaseService: DatabaseService
  ) {
    super();
  }

  public setDataModelerActionService(
    dataModelerService: DataModelerService
  ): void {
    this.dataModelerService = dataModelerService;
  }

  public setNotificationService(
    notificationService: NotificationService
  ): void {
    this.notificationService = notificationService;
  }

  public setDatabaseActionQueue(
    databaseActionQueue: ActionQueueOrchestrator<DatabaseActionsDefinition>
  ): void {
    this.databaseActionQueue = databaseActionQueue;
  }
}
