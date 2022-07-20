import type { DatabaseDataLoaderActions } from "$common/database-service/DatabaseDataLoaderActions";
import type { DatabaseTableActions } from "$common/database-service/DatabaseTableActions";
import type { DatabaseColumnActions } from "$common/database-service/DatabaseColumnActions";
import type { DuckDBClient } from "$common/database-service/DuckDBClient";
import type { DatabaseActions } from "$common/database-service/DatabaseActions";
import type {
  ActionServiceBase,
  ExtractActionTypeDefinitions,
  PickActionFunctions,
} from "$common/ServiceBase";
import { getActionMethods } from "$common/ServiceBase";
import type { DatabaseMetadata } from "$common/database-service/DatabaseMetadata";
import type { DatabaseMetricsExploreActions } from "$common/database-service/DatabaseMetricsExploreActions";
import type { DatabaseTimeSeriesActions } from "$common/database-service/DatabaseTimeSeriesActions";

type DatabaseActionsClasses = PickActionFunctions<
  DatabaseMetadata,
  DatabaseDataLoaderActions &
    DatabaseTableActions &
    DatabaseColumnActions &
    DatabaseMetricsExploreActions &
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
  ): Promise<any> {
    if (!this.actionsMap[action]?.[action]) {
      console.log(`${action} not found`);
      return;
    }
    const actionsInstance = this.actionsMap[action];
    return await actionsInstance[action].call(actionsInstance, null, ...args);
  }

  public async destroy(): Promise<void> {
    // FIXME add descriptive comment describing why this empy method is needed
  }
}
