import type { RootConfig } from "$common/config/RootConfig";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type { DuckDBClient } from "$common/database-service/DuckDBClient";
import { DataConnection } from "./DataConnection";

/**
 * Connects to an existing duck db.
 * Adds all existing table into the table states.
 * Periodically syncs
 */
export class DuckDbConnection extends DataConnection {
    public constructor(
        protected readonly config: RootConfig,
        protected readonly dataModelerService: DataModelerService,
        protected readonly dataModelerStateService: DataModelerStateService,
        protected readonly duckDbClient: DuckDBClient,
    ) {
        super(config, dataModelerService, dataModelerStateService);
    }

    public async init(): Promise<void> {
        const tables = await this.duckDbClient.execute("SHOW TABLES");
        for (const table of tables) {
            await this.dataModelerService.dispatch("addOrSyncTableFromDB", [table.name]);
        }
        console.log(`Imported tables: ${tables.map(table => table.name).join(", ")}`);
    }

    public async sync(): Promise<void> {
        // TODO
    }
}
