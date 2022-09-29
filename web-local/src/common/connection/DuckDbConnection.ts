import { TableSourceType } from "../../lib/types";
import type { RootConfig } from "../config/RootConfig";
import type { DataModelerService } from "../data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "../data-modeler-state-service/DataModelerStateService";
import {
  EntityType,
  StateType,
} from "../data-modeler-state-service/entity-state-service/EntityStateService";
import type { PersistentTableEntity } from "../data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import type { DuckDBClient } from "../database-service/DuckDBClient";
import { DataConnection } from "./DataConnection";

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
    // install HTTPFS extension, used to connect to remote sources like S3
    await this.duckDbClient.execute("INSTALL httpfs;", false, false);
    await this.duckDbClient.execute("LOAD httpfs;", false, false);

    if (this.config.database.databaseName === ":memory:") return;

    await this.sync();

    await this.dataModelerService.dispatch("loadModels", []);

    // this.syncTimer = setInterval(() => {
    //   this.sync();
    // }, DATABASE_POLLING_INTERVAL);
  }

  public async sync(): Promise<void> {
    const duckDbTables = await this.duckDbClient.execute<{
      table_name: string;
    }>(
      "SELECT table_name FROM information_schema.tables " +
        "WHERE table_type NOT ILIKE '%TEMPORARY' AND table_type NOT ILIKE '%VIEW';",
      false
    );
    const persistentTables = this.dataModelerStateService
      .getEntityStateService(EntityType.Table, StateType.Persistent)
      .getCurrentState().entities;
    const tablesFromLocalFiles = new Map<string, PersistentTableEntity>();
    const tablesFromSql = new Map<string, PersistentTableEntity>();
    const tablesFromDuckDb = new Map<string, PersistentTableEntity>();
    persistentTables.forEach((persistentTable) => {
      if (persistentTable.sourceType === TableSourceType.SQL) {
        tablesFromSql.set(persistentTable.name, persistentTable);
      }
      if (
        persistentTable.sourceType === TableSourceType.CSVFile ||
        persistentTable.sourceType === TableSourceType.ParquetFile
      ) {
        tablesFromLocalFiles.set(persistentTable.name, persistentTable);
      }
      if (persistentTable.sourceType === TableSourceType.DuckDB) {
        tablesFromDuckDb.set(persistentTable.name, persistentTable);
      }
    });
    for (const table of duckDbTables) {
      const tableName = table.table_name;

      if (tablesFromLocalFiles.has(tableName)) {
        await this.dataModelerService.dispatch("syncTable", [
          tablesFromLocalFiles.get(tableName).id,
        ]);
        tablesFromLocalFiles.delete(tableName);
      } else if (tablesFromDuckDb.has(tableName)) {
        await this.dataModelerService.dispatch("syncTable", [
          tablesFromDuckDb.get(tableName).id,
        ]);
      } else if (tablesFromSql.has(tableName)) {
        continue;
      } else {
        await this.dataModelerService.dispatch("addOrSyncTableFromDB", [
          tableName,
        ]);
      }
    }
    // clean up source entities for sources that don't exist in DuckDB anymore
    for (const removedTable of tablesFromLocalFiles.values()) {
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
