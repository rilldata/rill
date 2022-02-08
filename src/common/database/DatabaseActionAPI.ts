import type {DatabaseDataLoaderActions} from "$common/database/DatabaseDataLoaderActions";
import type {DatabaseTableActions} from "$common/database/DatabaseTableActions";
import type {DatabaseColumnActions} from "$common/database/DatabaseColumnActions";
import type {DuckDBClient} from "$common/database/DuckDBClient";
import type {DatabaseActions} from "$common/database/DatabaseActions";
import {getActionMethods} from "$common/ActionDispatcher";

export type PickDatabaseActionFunctions<Handler> = Pick<Handler, {
    [Action in keyof Handler]: Handler[Action] extends
        (...args: any[]) => Promise<any> ? Action : never
}[keyof Handler]>;
// TODO: why is this 'never' vs DataModelerStateManager & DataModelerActionAPI?
export type DatabaseActionsClassesAlias = (
    DatabaseDataLoaderActions &
    DatabaseTableActions &
    DatabaseColumnActions
)
export type DatabaseActionsClasses = PickDatabaseActionFunctions<
    DatabaseActionsClassesAlias
>;
// export type DatabaseActionsClasses
export type DatabaseActionsDefinition = {
    [Action in keyof DatabaseActionsClasses]: DatabaseActionsClasses[Action] extends
        (...args: infer Args) => Promise<any> ? Args : never
};

export class DatabaseActionAPI {
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
        // this.dispatch("")
    }

    // public async dispatch<Action extends keyof DatabaseActionsDefinition>(
    //     action: Action, args: DatabaseActionsDefinition[Action],
    // ): Promise<void> {
    //     if (!this.actionsMap[action]?.[action]) {
    //         console.log(`${action} not found`);
    //         return;
    //     }
    //     const actionsInstance = this.actionsMap[action];
    //     await actionsInstance[action].call(actionsInstance, this.dataModelerStateManager.getCurrentState(), ...args);
    // }

    public async destroy(): Promise<void> {}
}
