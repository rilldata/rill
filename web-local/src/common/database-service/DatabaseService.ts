import type { DatabaseActions } from "./DatabaseActions";
import type { DatabaseColumnActions } from "./DatabaseColumnActions";
import type { DatabaseDataLoaderActions } from "./DatabaseDataLoaderActions";
import type { DatabaseMetadata } from "./DatabaseMetadata";
import type { DatabaseMetricsExplorerActions } from "./DatabaseMetricsExplorerActions";
import type { DatabaseTableActions } from "./DatabaseTableActions";
import type { DatabaseTimeSeriesActions } from "./DatabaseTimeSeriesActions";
import type { DuckDBClient } from "./DuckDBClient";
import {
  ActionServiceBase,
  ExtractActionTypeDefinitions,
  getActionMethods,
  PickActionFunctions,
} from "../ServiceBase";

type DatabaseActionsClasses = PickActionFunctions<
  DatabaseMetadata,
  DatabaseDataLoaderActions &
    DatabaseTableActions &
    DatabaseColumnActions &
    DatabaseMetricsExplorerActions &
    DatabaseTimeSeriesActions
>;
export type DatabaseActionsDefinition = ExtractActionTypeDefinitions<
  DatabaseMetadata,
  DatabaseActionsClasses
>;

/**
 * Has actions that directly talk to the database.
 * Use dispatch for taking actions.
 *
 * Takes a databaseClient (Currently an instance of {@link DuckDBClient}
 * Also takes an array of {@link DatabaseActions} instances.
 * Actions supported is dependent on these instances passed in the constructor.
 * One caveat to note, type definition and actual instances passed might not match.
 */
export class DatabaseService
  implements ActionServiceBase<DatabaseActionsDefinition>
{
  private actionsMap: {
    [Action in keyof DatabaseActionsDefinition]?: DatabaseActionsClasses;
  } = {};

  public constructor(
    private readonly databaseClient: DuckDBClient,
    private readonly databaseActions: Array<DatabaseActions>
  ) {
    databaseActions.forEach((actions) => {
      getActionMethods(actions).forEach((action) => {
        this.actionsMap[action] = actions;
      });
    });
  }

  public async init(): Promise<void> {
    await this.databaseClient?.init();
  }

  public getDatabaseClient(): DuckDBClient {
    return this.databaseClient;
  }

  /**
   * Forwards action to the appropriate class.
   * @param action
   * @param args
   */
  public async dispatch<Action extends keyof DatabaseActionsDefinition>(
    action: Action,
    args: DatabaseActionsDefinition[Action]
  ): Promise<unknown> {
    if (!this.actionsMap[action]?.[action]) {
      console.log(`${action} not found`);
      return;
    }
    const actionsInstance = this.actionsMap[action];
    return await actionsInstance[action].call(actionsInstance, null, ...args);
  }

  public async destroy(): Promise<void> {
    await this.databaseClient?.destroy();
  }
}
