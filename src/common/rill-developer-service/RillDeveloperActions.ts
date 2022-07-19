import type { RootConfig } from "$common/config/RootConfig";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type { DatabaseService } from "$common/database-service/DatabaseService";
import type { ActionQueueOrchestrator } from "$common/priority-action-queue/ActionQueueOrchestrator";
import type { DatabaseActionsDefinition } from "$common/database-service/DatabaseService";
import type { RillDeveloperService } from "$common/rill-developer-service/RillDeveloperService";
import { ActionsBase } from "$common/ActionsBase";

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
