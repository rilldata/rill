import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type { TableActions } from "$common/data-modeler-service/TableActions";
import type {
  ExtractActionTypeDefinitions,
  PickActionFunctions,
} from "$common/ServiceBase";
import { getActionMethods } from "$common/ServiceBase";
import type { DataModelerActions } from "$common/data-modeler-service/DataModelerActions";
import type { ProfileColumnActions } from "$common/data-modeler-service/ProfileColumnActions";
import type { ModelActions } from "$common/data-modeler-service/ModelActions";
import type {
  DatabaseActionsDefinition,
  DatabaseService,
} from "$common/database-service/DatabaseService";
import type { NotificationService } from "$common/notifications/NotificationService";
import type {
  EntityRecord,
  EntityStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { ApplicationActions } from "$common/data-modeler-service/ApplicationActions";
import { ActionQueueOrchestrator } from "$common/priority-action-queue/ActionQueueOrchestrator";
import type { ActionResponse } from "$common/data-modeler-service/response/ActionResponse";
import { ActionResponseFactory } from "$common/data-modeler-service/response/ActionResponseFactory";
import { ActionDefinitionError } from "$common/errors/ActionDefinitionError";
import { ApplicationStatus } from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type {
  MetricsActionDefinition,
  MetricsService,
} from "$common/metrics-service/MetricsService";
import type {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

type DataModelerActionsClasses = PickActionFunctions<
  EntityStateActionArg<EntityRecord>,
  TableActions & ProfileColumnActions & ModelActions & ApplicationActions
>;
/**
 * Style definition for Rill Developer actions.
 * Action => [...args]
 */
export type DataModelerActionsDefinition = ExtractActionTypeDefinitions<
  EntityStateActionArg<EntityRecord>,
  DataModelerActionsClasses
>;

/**
 * Higher order / compound actions within Rill Developer that can call multiple state updates and other actions within Rill Developer Service
 * Maps 1-1 with actions taken by the interface, either a UI or CLI.
 * Examples: addModel, updateModelQuery etc.
 * Use dispatch for taking actions.
 *
 * Is passed an array {@link DataModelerActions} instances.
 * Actions supported is dependent on these instances passed in the constructor.
 * One caveat to note, type definition and actual instances passed might not match.
 */
export class DataModelerService {
  /**
   * Map of action to {@link DataModelerActions} instance.
   * This might not have an entry for everything in DataModelerActionsDefinition.
   * Depends on the {@link DataModelerActions} with which this class is instantiated.
   * @private
   */
  private actionsMap: {
    [Action in keyof DataModelerActionsDefinition]?: DataModelerActionsClasses;
  } = {};
  private runningCount = 0;

  public readonly databaseActionQueue: ActionQueueOrchestrator<DatabaseActionsDefinition>;

  public constructor(
    protected readonly dataModelerStateService: DataModelerStateService,
    private readonly databaseService: DatabaseService,
    private readonly notificationService: NotificationService,
    public readonly metricsService: MetricsService,
    private readonly dataModelerActions: Array<DataModelerActions>
  ) {
    this.databaseActionQueue =
      new ActionQueueOrchestrator<DatabaseActionsDefinition>(databaseService);
    dataModelerActions.forEach((actions) => {
      actions.setDataModelerActionService(this);
      actions.setNotificationService(notificationService);
      actions.setDatabaseActionQueue(this.databaseActionQueue);
      getActionMethods(actions).forEach((action) => {
        this.actionsMap[action] = actions;
      });
    });
  }

  public getDatabaseService(): DatabaseService {
    return this.databaseService;
  }

  public async init(): Promise<void> {
    await this.databaseService?.init();
  }

  /**
   * Forwards action to the appropriate class.
   * @param action
   * @param args
   */
  public async dispatch<Action extends keyof DataModelerActionsDefinition>(
    action: Action,
    args: DataModelerActionsDefinition[Action]
  ): Promise<ActionResponse> {
    if (!this.actionsMap[action]?.[action]) {
      return ActionResponseFactory.getErrorResponse(
        new ActionDefinitionError(`${action} not found`)
      );
    }
    const actionsInstance = this.actionsMap[action];
    const stateTypes = (
      actionsInstance?.constructor as typeof DataModelerActions
    ).actionToStateTypesMap[action];
    if (!stateTypes) {
      return ActionResponseFactory.getErrorResponse(
        new ActionDefinitionError(`No state types defined for ${action}`)
      );
    }
    if (this.runningCount === 0) {
      this.dataModelerStateService.dispatch("setApplicationStatus", [
        ApplicationStatus.Running,
      ]);
    }
    this.runningCount++;

    const stateService = this.dataModelerStateService.getEntityStateService(
      stateTypes[0] ?? (args[0] as EntityType),
      stateTypes[1] ?? (args[1] as StateType)
    );
    let returnResponse: ActionResponse;
    try {
      returnResponse = await actionsInstance[action].call(
        actionsInstance,
        { stateService },
        ...args
      );
      if (!returnResponse)
        returnResponse = ActionResponseFactory.getSuccessResponse();
    } catch (err) {
      returnResponse = ActionResponseFactory.getErrorResponse(err);
    }

    this.runningCount--;
    if (this.runningCount === 0) {
      this.dataModelerStateService.dispatch("setApplicationStatus", [
        ApplicationStatus.Idle,
      ]);
    }
    return returnResponse;
  }

  public async fireEvent<Event extends keyof MetricsActionDefinition>(
    event: Event,
    args: MetricsActionDefinition[Event]
  ): Promise<void> {
    await this.metricsService?.dispatch(event, args);
  }

  public async destroy(): Promise<void> {
    await this.databaseService?.destroy();
    await this.dataModelerStateService.destroy();
  }
}
