import {DataModelerStateManager} from "$common/state-actions/DataModelerStateManager";
import type {DataModelerActionAPI} from "$common/data-modeler-actions/DataModelerActionAPI";
import {DataModelerActionSocketAPI} from "$common/data-modeler-actions/DataModelerActionSocketAPI";

export function clientFactory(): {
    dataModelerStateManager: DataModelerStateManager,
    dataModelerActionAPI: DataModelerActionAPI,
} {
    const dataModelerStateManager = new DataModelerStateManager([]);
    const dataModelerActionAPI = new DataModelerActionSocketAPI(dataModelerStateManager,
        "http://localhost:3001");

    return {dataModelerStateManager, dataModelerActionAPI};
}
