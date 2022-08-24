import type { RootConfig } from "$common/config/RootConfig";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type { MetricsService } from "$common/metrics-service/MetricsService";
import type { RillDeveloperService } from "$common/rill-developer-service/RillDeveloperService";
import type { SocketNotificationService } from "$common/socket/SocketNotificationService";
import { FileActionsController } from "$server/controllers/FileActionsController";
import { MetricsDefinitionController } from "$server/controllers/MetricsDefinitionController";
import { MetricsDimensionController } from "$server/controllers/MetricsDimensionController";
import { MetricsMeasureController } from "$server/controllers/MetricsMeasureController";
import { SocketServer } from "$server/SocketServer";
import bodyParser from "body-parser";
import cors from "cors";
import express from "express";
import fileUpload from "express-fileupload";
import { existsSync, mkdirSync } from "fs";
import http from "http";
import { MetricViewController } from "$server/controllers/MetricViewController";

const STATIC_FILES = `${__dirname}/../../build`;

export class ExpressServer {
  public readonly app: express.Application;
  private readonly server: http.Server;
  private readonly socketServer: SocketServer;

  constructor(
    private readonly config: RootConfig,
    private readonly dataModelerService: DataModelerService,
    private readonly rillDeveloperService: RillDeveloperService,
    dataModelerStateService: DataModelerStateService,
    notificationService: SocketNotificationService,
    metricsService: MetricsService
  ) {
    this.app = express();
    this.server = http.createServer(this.app);

    this.setupMiddlewares();
    this.setupControllers();

    this.socketServer = new SocketServer(
      config,
      dataModelerService,
      dataModelerStateService,
      metricsService,
      this.server
    );
    notificationService.setSocketServer(this.socketServer.getSocketServer());

    if (config.server.serveStaticFile) {
      this.app.use(express.static(STATIC_FILES));
    }
  }

  public async init(): Promise<void> {
    await this.socketServer.init();
    this.server.listen(this.config.server.serverPort);
    console.log(`Server started at ${this.config.server.serverUrl}`);
  }

  public async destroy(): Promise<void> {
    await this.socketServer.destroy();
    this.server.close();
  }

  private setupMiddlewares() {
    this.app.use(
      cors({
        origin: this.config.server.uiUrl,
      })
    );
    this.app.use(bodyParser.json());

    const tmpFolder = `${this.config.projectFolder}/tmp`;
    if (!existsSync(tmpFolder)) mkdirSync(tmpFolder);
    this.app.use(
      fileUpload({
        useTempFiles: true,
        tempFileDir: tmpFolder,
      })
    );
  }

  private setupControllers() {
    new FileActionsController(
      this.config,
      this.dataModelerService,
      this.rillDeveloperService
    ).setup(this.app, "/api/file");
    if (!this.rillDeveloperService) return;

    [
      MetricsDefinitionController,
      MetricsDimensionController,
      MetricsMeasureController,
      MetricViewController,
    ].forEach((MetricsControllerClass) =>
      new MetricsControllerClass(
        this.config,
        this.dataModelerService,
        this.rillDeveloperService
      ).setup(this.app, "/api")
    );
  }
}
