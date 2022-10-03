import type { RootConfig } from "../config/RootConfig";
import type { DataModelerStateService } from "../data-modeler-state-service/DataModelerStateService";
import type { DatabaseService } from "../database-service/DatabaseService";
import type { ActionQueueOrchestrator } from "../priority-action-queue/ActionQueueOrchestrator";
import type { DatabaseActionsDefinition } from "../database-service/DatabaseService";
import type { RillDeveloperService } from "./RillDeveloperService";
import { ActionsBase } from "../ActionsBase";

export class RillDeveloperActions extends ActionsBase {
  protected rillDeveloperService: RillDeveloperService;
  protected databaseActionQueue: ActionQueueOrchestrator<DatabaseActionsDefinition>;

  constructor(
    protected readonly config: RootConfig,
    protected readonly dataModelerStateService: DataModelerStateService,
    protected readonly databaseService: DatabaseService
  ) {
    super();
  }

  public setRillDeveloperService(
    rillDeveloperService: RillDeveloperService
  ): void {
    this.rillDeveloperService = rillDeveloperService;
  }

  public setDatabaseActionQueue(
    databaseActionQueue: ActionQueueOrchestrator<DatabaseActionsDefinition>
  ): void {
    this.databaseActionQueue = databaseActionQueue;
  }
}
