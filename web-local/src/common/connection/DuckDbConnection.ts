import type { RootConfig } from "../config/RootConfig";
import type { DataModelerService } from "../data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "../data-modeler-state-service/DataModelerStateService";
import {
  EntityType,
  StateType,
} from "../data-modeler-state-service/entity-state-service/EntityStateService";
import type { DuckDBClient } from "../database-service/DuckDBClient";
import { DataConnection } from "./DataConnection";
import type { PersistentTableEntity } from "../data-modeler-state-service/entity-state-service/PersistentTableEntityService";

/**
 * Connects to an existing duck db.
 * Adds all existing table into the table states.
 * Periodically syncs
 */
export class DuckDbConnection extends DataConnection {
  private syncTimer: NodeJS.Timer;

  public constructor(
    protected readonly config: RootConfig,
    protected readonly dataModelerService: DataModelerService,
    protected readonly dataModelerStateService: DataModelerStateService,
    protected readonly duckDbClient: DuckDBClient
  ) {
    super(config, dataModelerService, dataModelerStateService);
  }

  public async init(): Promise<void> {
    if (this.config.database.databaseName === ":memory:") return;

    await this.sync();

    await this.dataModelerService.dispatch("loadModels", []);

    // this.syncTimer = setInterval(() => {
    //   this.sync();
    // }, DATABASE_POLLING_INTERVAL);
  }

  public async sync(): Promise<void> {
    const tables = await this.duckDbClient.execute<{ table_name: string }>(
      "SELECT table_name FROM information_schema.tables " +
        "WHERE table_type NOT ILIKE '%TEMPORARY' AND table_type NOT ILIKE '%VIEW';",
      false
    );
    const persistentTables = this.dataModelerStateService
      .getEntityStateService(EntityType.Table, StateType.Persistent)
      .getCurrentState().entities;

    const existingTables = new Map<string, PersistentTableEntity>();
    persistentTables.forEach((persistentTable) =>
      existingTables.set(persistentTable.tableName, persistentTable)
    );

    for (const table of tables) {
      const tableName = table.table_name;
      if (!tableName || tableName.endsWith("___")) continue;
      if (existingTables.has(tableName)) {
        await this.dataModelerService.dispatch("syncTable", [
          existingTables.get(tableName).id,
        ]);
        existingTables.delete(tableName);
      } else {
        await this.dataModelerService.dispatch("addOrSyncTableFromDB", [
          tableName,
        ]);
      }
    }
    for (const removedTable of existingTables.values()) {
      await this.dataModelerService.dispatch("dropTable", [
        removedTable.tableName,
        true,
      ]);
    }
  }

  public async destroy(): Promise<void> {
    if (this.syncTimer) clearInterval(this.syncTimer);
  }
}
