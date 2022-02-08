import {DataModelerStateService} from "$common/data-modeler-state-service/DataModelerStateService";
import type {DataModelerService} from "$common/data-modeler-service/DataModelerService";
import {DataModelerSocketService} from "$common/data-modeler-service/DataModelerSocketService";

export function clientFactory(): {
    dataModelerStateService: DataModelerStateService,
    dataModelerService: DataModelerService,
} {
    const dataModelerStateService = new DataModelerStateService([]);
    const dataModelerService = new DataModelerSocketService(dataModelerStateService,
        "http://localhost:3001");

    return {dataModelerStateService, dataModelerService};
}
