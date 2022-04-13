import type { RootConfig } from "$common/config/RootConfig";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";

export abstract class DataConnection {
    protected constructor(
        protected readonly config: RootConfig,
        protected readonly dataModelerService: DataModelerService,
        protected readonly dataModelerStateService: DataModelerStateService,
    ) {}

    public abstract init(): Promise<void>;
    public abstract sync(): Promise<void>;
}
