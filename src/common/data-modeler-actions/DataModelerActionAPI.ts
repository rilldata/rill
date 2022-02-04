import type {DataModelerStateManager} from "$common/state-actions/DataModelerStateManager";
import type {DatasetActions} from "$common/data-modeler-actions/DatasetActions";
import type {ExtractActionTypeDefinitions, PickActionFunctions} from "$common/ActionDispatcher";
import type {DuckDBClient} from "$common/duckdb/DuckDBClient";
import type {DataModelerActions} from "$common/data-modeler-actions/DataModelerActions";
import type {ProfileColumnActions} from "$common/data-modeler-actions/ProfileColumnActions";
import {getActionMethods} from "$common/ActionDispatcher";

export type DataModelerActionsClasses = PickActionFunctions<DatasetActions & ProfileColumnActions>;
export type DataModelerActionsDefinition = ExtractActionTypeDefinitions<DataModelerActionsClasses>;

export class DataModelerActionAPI {
    private actionsMap: {
        [Action in keyof DataModelerActionsDefinition]?: DataModelerActionsClasses
    } = {};

    public constructor(protected readonly dataModelerStateManager: DataModelerStateManager,
                       private readonly duckDBClient: DuckDBClient,
                       private readonly dataModelerActions: Array<DataModelerActions>) {
        dataModelerActions.forEach((actions) => {
            actions.setDataModelerActionAPI(this);
            getActionMethods(actions).forEach(action => {
                this.actionsMap[action] = actions;
            });
        });
    }

    public async init(): Promise<void> {
        this.dataModelerStateManager.init();
        await this.duckDBClient?.init();
    }

    public async dispatch<Action extends keyof DataModelerActionsDefinition>(
        action: Action, args: DataModelerActionsDefinition[Action],
    ): Promise<void> {
        if (!this.actionsMap[action]?.[action]) {
            console.log(`${action} not found`);
            return;
        }
        const actionsInstance = this.actionsMap[action];
        await actionsInstance[action].call(actionsInstance, this.dataModelerStateManager.getCurrentState(), ...args);
    }

    public async destroy(): Promise<void> {}
}
