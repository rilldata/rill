import {DataModelerStateService} from "$common/state-actions/DataModelerStateService";
import type {DataModelerService} from "$common/data-modeler-actions/DataModelerService";
import {DataModelerSocketService} from "$common/data-modeler-actions/DataModelerSocketService";

export function clientFactory(): {
    dataModelerStateService: DataModelerStateService,
    dataModelerService: DataModelerService,
} {
    const dataModelerStateService = new DataModelerStateService([]);
    const dataModelerService = new DataModelerSocketService(dataModelerStateService,
        "http://localhost:3001");

    return {dataModelerStateService, dataModelerService};
}
