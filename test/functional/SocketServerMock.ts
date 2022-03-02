import type {DataModelerService} from "$common/data-modeler-service/DataModelerService";
import type {DataModelerStateService} from "$common/data-modeler-state-service/DataModelerStateService";
import type {DataModelerSocketServiceMock} from "./DataModelerSocketServiceMock";
import type { RootConfig } from "$common/config/RootConfig";
import {
    DataModelerStateSyncService
} from "$common/data-modeler-state-service/sync-service/DataModelerStateSyncService";

export class SocketServerMock {
    private readonly dataModelerStateSyncService: DataModelerStateSyncService;

    constructor(private readonly config: RootConfig,
                private readonly dataModelerService: DataModelerService,
                private readonly dataModelerStateService: DataModelerStateService,
                private readonly dataModelerSocketServiceMock: DataModelerSocketServiceMock) {
        this.dataModelerStateSyncService = new DataModelerStateSyncService(
            config, dataModelerStateService.entityStateServices,
            dataModelerService, dataModelerStateService);
    }

    public async init(): Promise<void> {
        await this.dataModelerService.init();
        await this.dataModelerStateSyncService.init();

        this.dataModelerStateService.subscribePatches((entityType, stateType, patches) => {
            this.dataModelerSocketServiceMock.applyPatches(entityType, stateType, patches);
        });

        this.dataModelerSocketServiceMock.initialState(this.dataModelerStateService.getCurrentStates());
    }

    public async dispatch(action: string, args: Array<any>) {
        return this.dataModelerService.dispatch(action as any, args);
    }

    public async destroy(): Promise<void> {
        await this.dataModelerStateSyncService.destroy();
    }
}
