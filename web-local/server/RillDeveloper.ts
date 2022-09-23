import { existsSync } from "fs";
import type { RootConfig } from "../common/config/RootConfig";
import { DuckDbConnection } from "../common/connection/DuckDbConnection";
import type { DataModelerService } from "../common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "../common/data-modeler-state-service/DataModelerStateService";
import { DataModelerStateSyncService } from "../common/data-modeler-state-service/sync-service/DataModelerStateSyncService";
import type { MetricsService } from "../common/metrics-service/MetricsService";
import type { NotificationService } from "../common/notifications/NotificationService";
import { dataModelerServiceFactory } from "./serverFactory";

/**
 * Wrapper class that initializes other classes.
 * To be used on the server only.
 */
export class RillDeveloper {
  private readonly duckDbConnection: DuckDbConnection;
  public constructor(
    public readonly config: RootConfig,
    public readonly dataModelerService: DataModelerService,
    public readonly dataModelerStateService: DataModelerStateService,
    public readonly dataModelerStateSyncService: DataModelerStateSyncService,
    public readonly metricsService: MetricsService,
    public readonly notificationService: NotificationService
  ) {
    this.duckDbConnection = new DuckDbConnection(
      this.config,
      this.dataModelerService,
      this.dataModelerStateService,
      this.dataModelerService.getDatabaseService().getDatabaseClient()
    );
  }

  public async init(): Promise<void> {
    const alreadyInitialized = existsSync(this.config.state.stateFolder);

    await this.dataModelerStateSyncService.init();
    if (alreadyInitialized) {
      this.config.project.duckDbPath =
        this.dataModelerStateService.getApplicationState().duckDbPath;
    }
    if (this.config.project.duckDbPath) {
      this.config.database.databaseName = this.config.project.duckDbPath;
    }

    await this.dataModelerService.init();
    if (!alreadyInitialized && this.config.project.duckDbPath) {
      this.dataModelerStateService.dispatch("setDuckDbPath", [
        this.config.project.duckDbPath,
      ]);
    }
    await this.duckDbConnection.init();
  }

  public async destroy() {
    await this.dataModelerStateSyncService.destroy();
    await this.duckDbConnection.destroy();
    await this.dataModelerService.destroy();
  }

  public static getRillDeveloper(config: RootConfig) {
    const {
      dataModelerService,
      dataModelerStateService,
      metricsService,
      notificationService,
    } = dataModelerServiceFactory(config);

    const dataModelerStateSyncService = new DataModelerStateSyncService(
      config,
      dataModelerStateService.entityStateServices,
      dataModelerService,
      dataModelerStateService
    );

    return new RillDeveloper(
      config,
      dataModelerService,
      dataModelerStateService,
      dataModelerStateSyncService,
      metricsService,
      notificationService
    );
  }
}
