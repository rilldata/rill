import type {DataModelerService} from "$common/data-modeler-service/DataModelerService";
import type {DataModelerStateService} from "$common/data-modeler-state-service/DataModelerStateService";
import type {DataModelerSocketServiceMock} from "./DataModelerSocketServiceMock";

export class SocketServerMock {
    constructor(private readonly dataModelerService: DataModelerService,
                private readonly dataModelerStateService: DataModelerStateService,
                private readonly dataModelerSocketServiceMock: DataModelerSocketServiceMock) {}

    public async init(): Promise<void> {
        await this.dataModelerService.init();

        this.dataModelerStateService.subscribePatches((entityType, stateType, patches) => {
            this.dataModelerSocketServiceMock.applyPatches(entityType, stateType, patches);
        });

        this.dataModelerSocketServiceMock.initialState(this.dataModelerStateService.getCurrentStates());
    }

    public async dispatch(action: string, args: Array<any>) {
        return this.dataModelerService.dispatch(action as any, args);
    }
}
