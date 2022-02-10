import type {DataModelerStateService} from "$common/data-modeler-state-service/DataModelerStateService";
import type {DatasetActions} from "$common/data-modeler-service/DatasetActions";
import type {ExtractActionTypeDefinitions, PickActionFunctions} from "$common/ServiceBase";
import type {DataModelerActions} from "$common/data-modeler-service/DataModelerActions";
import type {ProfileColumnActions} from "$common/data-modeler-service/ProfileColumnActions";
import type {ModelActions} from "$common/data-modeler-service/ModelActions";
import {getActionMethods} from "$common/ServiceBase";
import {IDLE_STATUS, RUNNING_STATUS} from "$common/constants";
import type {DataModelerState} from "$lib/types";
import type {DatabaseService} from "$common/database-service/DatabaseService";

export type DataModelerActionsClasses = PickActionFunctions<DataModelerState, (
    DatasetActions &
    ProfileColumnActions &
    ModelActions
)>;
export type DataModelerActionsDefinition = ExtractActionTypeDefinitions<DataModelerState, DataModelerActionsClasses>;

export class DataModelerService {
    private actionsMap: {
        [Action in keyof DataModelerActionsDefinition]?: DataModelerActionsClasses
    } = {};

    public constructor(protected readonly dataModelerStateService: DataModelerStateService,
                       private readonly databaseService: DatabaseService,
                       private readonly dataModelerActions: Array<DataModelerActions>) {
        dataModelerActions.forEach((actions) => {
            actions.setDataModelerActionAPI(this);
            getActionMethods(actions).forEach(action => {
                this.actionsMap[action] = actions;
            });
        });
    }

    public async init(): Promise<void> {
        this.dataModelerStateService.init();
        await this.databaseService?.init();
    }

    public async dispatch<Action extends keyof DataModelerActionsDefinition>(
        action: Action, args: DataModelerActionsDefinition[Action],
    ): Promise<void> {
        if (!this.actionsMap[action]?.[action]) {
            console.log(`${action} not found`);
            return;
        }
        const actionsInstance = this.actionsMap[action];
        // this.dataModelerStateService.dispatch("setStatus", [RUNNING_STATUS]);
        await actionsInstance[action].call(actionsInstance, this.dataModelerStateService.getCurrentState(), ...args);
        // this.dataModelerStateService.dispatch("setStatus", [IDLE_STATUS]);
    }

    public async destroy(): Promise<void> {}
}
