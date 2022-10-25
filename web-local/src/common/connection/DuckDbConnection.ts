import { DATABASE_POLLING_INTERVAL } from "@rilldata/web-local/common/constants";
import { getMapFromArray } from "@rilldata/web-local/common/utils/arrayUtils";
import { runtimeServiceListCatalogObjects } from "web-common/src/runtime-client";
import type { RootConfig } from "../config/RootConfig";
import type { DataModelerService } from "../data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "../data-modeler-state-service/DataModelerStateService";
import {
  EntityType,
  StateType,
} from "../data-modeler-state-service/entity-state-service/EntityStateService";
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
    if (this.config.database.databaseName === ":memory:") return;

    await this.sync();

    await this.dataModelerService.dispatch("loadModels", []);

    this.syncTimer = setInterval(() => {
      this.sync();
    }, DATABASE_POLLING_INTERVAL);
  }

  public async sync(): Promise<void> {
    const catalogs = await runtimeServiceListCatalogObjects(
      this.duckDbClient.getInstanceId()
    );
    const catalogsMap = getMapFromArray(
      catalogs.objects,
      (object) => object.source?.name
    );
    const persistentTables = this.dataModelerStateService
      .getEntityStateService(EntityType.Table, StateType.Persistent)
      .getCurrentState().entities;

    for (const persistentTable of persistentTables) {
      if (
        persistentTable.previousTableName &&
        catalogsMap.has(persistentTable.previousTableName)
      ) {
        // hack to process renames.
        // we set previousTableName when rename is triggered
        // here we get confirmation that rename finished
        this.dataModelerStateService.dispatch("updateTableName", [
          persistentTable.id,
          persistentTable.previousTableName,
        ]);
        catalogsMap.delete(persistentTable.previousTableName);
        continue;
      }

      if (!catalogsMap.has(persistentTable.tableName)) {
        await this.dataModelerService.dispatch("dropTable", [
          persistentTable.tableName,
        ]);
        continue;
      }

      catalogsMap.delete(persistentTable.tableName);
      await this.dataModelerService.dispatch("syncTable", [persistentTable.id]);
    }

    for (const absentCatalogs of catalogsMap.values()) {
      if (!absentCatalogs.source) continue;
      await this.dataModelerService.dispatch("addOrSyncTableFromDB", [
        absentCatalogs.source.name,
      ]);
    }
  }

  public async destroy(): Promise<void> {
    if (this.syncTimer) clearInterval(this.syncTimer);
  }
}
