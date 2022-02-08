import {DataModelerActionsDefinition, DataModelerService} from "$common/data-modeler-actions/DataModelerService";
import type {DataModelerStateService} from "$common/state-actions/DataModelerStateService";
import type {SocketServerMock} from "./SocketServerMock";
import type {DataModelerState} from "$lib/types";
import type {Patch} from "immer";

export class DataModelerSocketServiceMock extends DataModelerService {
    public socketServerMock: SocketServerMock;

    public constructor(dataModelerStateService: DataModelerStateService) {
        super(dataModelerStateService, null, []);
    }

    public async init(): Promise<void> {
        this.dataModelerStateService.init();
    }

    public initialState(state: DataModelerState) {
        this.dataModelerStateService.updateState(state);
    }

    public applyPatches(patches: Array<Patch>) {
        this.dataModelerStateService.applyPatches(patches);
    }

    public async dispatch<Action extends keyof DataModelerActionsDefinition>(
        action: Action, args: DataModelerActionsDefinition[Action],
    ): Promise<void> {
        return this.socketServerMock.dispatch(action, args);
    }
}
