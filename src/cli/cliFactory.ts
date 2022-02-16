import { dataModelerServiceFactory } from "$common/serverFactory";
import { RootConfig } from "$common/config/RootConfig";
import { DatabaseConfig } from "$common/config/DatabaseConfig";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type { DataModelerState } from "$lib/types";
import { existsSync, readFileSync } from "fs";

export const SAVED_STATE_FILE = "saved-state.json";
export const DATABASE_NAME = "stage.db";

export async function getCliInstances(path: string): Promise<{
    dataModelerService: DataModelerService,
    dataModelerStateService: DataModelerStateService,
}> {
    let initialState: DataModelerState;
    if (existsSync(`${path}/${SAVED_STATE_FILE}`)) {
        initialState = JSON.parse(readFileSync(`${path}/${SAVED_STATE_FILE}`).toString());
    }
    const {dataModelerService, dataModelerStateService} = dataModelerServiceFactory(new RootConfig({
        database: new DatabaseConfig({ databaseName: `${path}/${DATABASE_NAME}` }),
    }));
    await dataModelerService.init(initialState);
    return {dataModelerService, dataModelerStateService};
}
