import type { RootConfig } from "$common/config/RootConfig";
import { DATABASE_POLLING_INTERVAL } from "$common/constants";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { DuckDBClient } from "$common/database-service/DuckDBClient";
import { DataConnection } from "./DataConnection";
import type { PersistentSourceEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentSourceEntityService";

/**
 * Connects to an existing duck db.
 * Adds all existing tables into the source states.
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
    await this.sync();

    this.syncTimer = setInterval(() => {
      this.sync();
    }, DATABASE_POLLING_INTERVAL);
  }

  public async sync(): Promise<void> {
    const tables = await this.duckDbClient.execute(
      "SELECT table_name FROM information_schema.tables " +
        "WHERE table_type NOT ILIKE '%TEMPORARY' AND table_type NOT ILIKE '%VIEW';"
    );
    const persistentSources = this.dataModelerStateService
      .getEntityStateService(EntityType.Source, StateType.Persistent)
      .getCurrentState().entities;

    const existingSources = new Map<string, PersistentSourceEntity>();
    persistentSources.forEach((persistentSource) =>
      existingSources.set(persistentSource.sourceName, persistentSource)
    );

    for (const table of tables) {
      const sourceName = table.table_name;
      if (existingSources.has(sourceName)) {
        await this.dataModelerService.dispatch("syncSource", [
          existingSources.get(sourceName).id,
        ]);
        existingSources.delete(sourceName);
      } else {
        await this.dataModelerService.dispatch("addOrSyncSourceFromDB", [
          sourceName,
        ]);
      }
    }
    for (const removedSource of existingSources.values()) {
      await this.dataModelerService.dispatch("dropSource", [
        removedSource.sourceName,
        true,
      ]);
    }
  }

  public async destroy(): Promise<void> {
    if (this.syncTimer) clearInterval(this.syncTimer);
  }
}
