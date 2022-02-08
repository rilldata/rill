import type {DatabaseDataLoaderActions} from "$common/database/DatabaseDataLoaderActions";
import type {DatabaseTableActions} from "$common/database/DatabaseTableActions";
import type {DatabaseColumnActions} from "$common/database/DatabaseColumnActions";
import type {DuckDBClient} from "$common/database/DuckDBClient";
import type {DatabaseActions} from "$common/database/DatabaseActions";
import {ExtractActionTypeDefinitions, getActionMethods, PickActionFunctions} from "$common/ServiceBase";
import type {DatabaseMetadata} from "$common/database/DatabaseMetadata";

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

    public constructor(private readonly duckDBClient: DuckDBClient,
                       private readonly databaseActions: Array<DatabaseActions>) {
        databaseActions.forEach((actions) => {
            getActionMethods(actions).forEach(action => {
                this.actionsMap[action] = actions;
            });
        });
    }

    public async init(): Promise<void> {
        await this.duckDBClient?.init();
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
