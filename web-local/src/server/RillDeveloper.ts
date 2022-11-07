import type { RootConfig } from "@rilldata/web-local/common/config/RootConfig";
import { DuckDbConnection } from "@rilldata/web-local/common/connection/DuckDbConnection";
import type { DataModelerService } from "@rilldata/web-local/common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "@rilldata/web-local/common/data-modeler-state-service/DataModelerStateService";
import {
  EntityType,
  StateType,
} from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { DataModelerStateSyncService } from "@rilldata/web-local/common/data-modeler-state-service/sync-service/DataModelerStateSyncService";
import type { MetricsService } from "@rilldata/web-local/common/metrics-service/MetricsService";
import type { NotificationService } from "@rilldata/web-local/common/notifications/NotificationService";
import axios from "axios";
import { existsSync, mkdirSync } from "fs";
import type {
  V1CreateRepoRequest,
  V1CreateRepoResponse,
} from "@rilldata/web-common/runtime-client";
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

  public async init(): Promise<void> {
    mkdirSync(this.config.projectFolder, {
      recursive: true,
    });
    const alreadyInitialized = existsSync(this.config.state.stateFolder);

    // this essentially calls DuckdbClient.init. hence moving it to the beginning
    if (alreadyInitialized) {
      this.config.project.duckDbPath =
        this.dataModelerStateService.getApplicationState().duckDbPath;
    }
    if (this.config.project.duckDbPath) {
      this.config.database.databaseName = this.config.project.duckDbPath;
    }
    await this.dataModelerService.init();
    await this.dataModelerStateSyncService.init();

    if (!alreadyInitialized && this.config.project.duckDbPath) {
      this.dataModelerStateService.dispatch("setDuckDbPath", [
        this.config.project.duckDbPath,
      ]);
    }
    await this.duckDbConnection.init();

    await this.createRepo();
  }

  public async destroy() {
    await this.dataModelerStateSyncService.destroy();
    await this.duckDbConnection.destroy();
    await this.dataModelerService.destroy();
  }

  // temporary hack to create a repo for the project folder
  private async createRepo() {
    const resp = await axios.post(
      `${this.config.database.runtimeUrl}/v1/repos`,
      {
        driver: "file",
        dsn: this.config.projectFolder,
      } as V1CreateRepoRequest
    );
    const repoResp: V1CreateRepoResponse = resp.data;
    this.dataModelerStateService
      .getEntityStateService(EntityType.Application, StateType.Derived)
      .updateState(
        (draft) => {
          draft.repoId = repoResp.repo.repoId;
        },
        () => {
          // no-op
        }
      );
  }
}
