import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type { DatabaseService } from "$common/database-service/DatabaseService";
import type { RillDeveloperActions } from "$common/rill-developer-service/RillDeveloperActions";
import type { PickActionFunctions } from "$common/ServiceBase";
import type { RillRequestContext } from "$common/rill-developer-service/RillRequestContext";
import type { MetricsDefinitionActions } from "$common/rill-developer-service/MetricsDefinitionActions";
import type { DimensionsActions } from "$common/rill-developer-service/DimensionsActions";
import type { MeasuresActions } from "$common/rill-developer-service/MeasuresActions";
import type { ExtractActionTypeDefinitions } from "$common/ServiceBase";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import { getActionMethods } from "$common/ServiceBase";
import type { ActionResponse } from "$common/data-modeler-service/response/ActionResponse";
import { ActionResponseFactory } from "$common/data-modeler-service/response/ActionResponseFactory";
import { ActionDefinitionError } from "$common/errors/ActionDefinitionError";

type RillDeveloperActionsClasses = PickActionFunctions<
  RillRequestContext<any, any>,
  MetricsDefinitionActions & DimensionsActions & MeasuresActions
>;
export type RillDeveloperActionsDefinition = ExtractActionTypeDefinitions<
  RillRequestContext<any, any>,
  RillDeveloperActionsClasses
>;

/**
 * This is mostly a copy of DataModelerService but renamed to be match the product.
 * It also has RillRequestContext that passes on data.
 */
export class RillDeveloperService {
  private actionsMap: {
    [Action in keyof RillDeveloperActionsDefinition]?: RillDeveloperActionsClasses;
  } = {};

  public constructor(
    protected readonly dataModelerStateService: DataModelerStateService,
    private readonly dataModelerService: DataModelerService,
    private readonly databaseService: DatabaseService,
    private readonly rillDeveloperActions: Array<RillDeveloperActions>
  ) {
    rillDeveloperActions.forEach((actions) => {
      actions.setRillDeveloperService(this);
      actions.setDatabaseActionQueue(dataModelerService.databaseActionQueue);
      getActionMethods(actions).forEach((action) => {
        this.actionsMap[action] = actions;
      });
    });
  }

  public async dispatch<Action extends keyof RillDeveloperActionsDefinition>(
    context: RillRequestContext<any, any>,
    action: Action,
    args: RillDeveloperActionsDefinition[Action]
  ): Promise<ActionResponse> {
    if (!this.actionsMap[action]?.[action]) {
      return ActionResponseFactory.getErrorResponse(
        new ActionDefinitionError(`${action} not found`)
      );
    }
    const actionsInstance = this.actionsMap[action];
    let returnResponse: ActionResponse;
    try {
      returnResponse = await actionsInstance[action].call(
        actionsInstance,
        context,
        ...args
      );
      if (!returnResponse)
        returnResponse = ActionResponseFactory.getSuccessResponse();
    } catch (err) {
      returnResponse = ActionResponseFactory.getErrorResponse(err);
    }

    return returnResponse;
  }
}
