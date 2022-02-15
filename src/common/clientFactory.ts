import {DataModelerStateService} from "$common/data-modeler-state-service/DataModelerStateService";
import type {DataModelerService} from "$common/data-modeler-service/DataModelerService";
import {DataModelerSocketService} from "$common/socket/DataModelerSocketService";
import type {RootConfig} from "$common/config/RootConfig";

export function clientFactory(config: RootConfig): {
    dataModelerStateService: DataModelerStateService,
    dataModelerService: DataModelerService,
} {
    const dataModelerStateService = new DataModelerStateService([]);
    const dataModelerService = new DataModelerSocketService(dataModelerStateService, config.server);

    return {dataModelerStateService, dataModelerService};
}
