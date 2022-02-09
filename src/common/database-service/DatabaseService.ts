import type {DatabaseDataLoaderActions} from "$common/database-service/DatabaseDataLoaderActions";
import type {DatabaseTableActions} from "$common/database-service/DatabaseTableActions";
import type {DatabaseColumnActions} from "$common/database-service/DatabaseColumnActions";
import type {DuckDBClient} from "$common/database-service/DuckDBClient";
import type {DatabaseActions} from "$common/database-service/DatabaseActions";
import {ExtractActionTypeDefinitions, getActionMethods, PickActionFunctions} from "$common/ServiceBase";
import type {DatabaseMetadata} from "$common/database-service/DatabaseMetadata";

export type DatabaseActionsClasses = PickActionFunctions<DatabaseMetadata, (
    DatabaseDataLoaderActions &
    DatabaseTableActions &
    DatabaseColumnActions
)>;
export type DatabaseActionsDefinition = ExtractActionTypeDefinitions<DatabaseMetadata, DatabaseActionsClasses>;

export class DatabaseService {
    private actionsMap: {
        [Action in keyof DatabaseActionsDefinition]?: DatabaseActionsClasses
    } = {};

    public constructor(private readonly databaseClient: DuckDBClient,
                       private readonly databaseActions: Array<DatabaseActions>) {
        databaseActions.forEach((actions) => {
            getActionMethods(actions).forEach(action => {
                this.actionsMap[action] = actions;
            });
        });
    }

    public async init(): Promise<void> {
        await this.databaseClient?.init();
    }

    public async dispatch<Action extends keyof DatabaseActionsDefinition>(
        action: Action, args: DatabaseActionsDefinition[Action],
    ): Promise<any> {
        if (!this.actionsMap[action]?.[action]) {
            console.log(`${action} not found`);
            return;
        }
        const actionsInstance = this.actionsMap[action];
        return await actionsInstance[action].call(actionsInstance, null, ...args);
    }

    public async destroy(): Promise<void> {}
}
