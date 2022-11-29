import type { RootConfig } from "@rilldata/web-local/common/config/RootConfig";
import type { DataModelerService } from "@rilldata/web-local/common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "@rilldata/web-local/common/data-modeler-state-service/DataModelerStateService";
import type { RillDeveloperService } from "@rilldata/web-local/common/rill-developer-service/RillDeveloperService";
import type { SocketNotificationService } from "@rilldata/web-local/common/socket/SocketNotificationService";
import { FileActionsController } from "./controllers/FileActionsController";
import { MetricsDefinitionController } from "./controllers/MetricsDefinitionController";
import { MetricsDimensionController } from "./controllers/MetricsDimensionController";
import { MetricsMeasureController } from "./controllers/MetricsMeasureController";
import { MetricsViewController } from "./controllers/MetricsViewController";
import { SocketServer } from "./SocketServer";
import bodyParser from "body-parser";
import cors from "cors";
import express, { Request, Response } from "express";
import fileUpload from "express-fileupload";
import { existsSync, mkdirSync } from "fs";
import http from "http";
import path from "path";

const STATIC_FILES = existsSync(`${__dirname}/../../build`)
  ? `${__dirname}/../../build`
  : `${__dirname}/../../../../build`;
const SVELTEKIT_FALLBACK_PAGE = "index.html";

export class ExpressServer {
  public readonly app: express.Application;
  private readonly server: http.Server;
  private readonly socketServer: SocketServer;

  constructor(
    private readonly config: RootConfig,
    private readonly dataModelerService: DataModelerService,
    private readonly rillDeveloperService: RillDeveloperService,
    dataModelerStateService: DataModelerStateService,
    notificationService: SocketNotificationService
  ) {
    this.app = express();
    this.server = http.createServer(this.app);

    this.setupMiddlewares();
    this.setupControllers();

    this.socketServer = new SocketServer(
      config,
      dataModelerService,
      dataModelerStateService,
      this.server
    );
    notificationService.setSocketServer(this.socketServer.getSocketServer());

    if (config.server.serveStaticFile) {
      this.app.use(express.static(STATIC_FILES));
    }

    // add fallback route
    this.app.get("*", (req, res) => {
      res.sendFile(path.resolve(STATIC_FILES, SVELTEKIT_FALLBACK_PAGE));
    });
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
      MetricsViewController,
    ].forEach((MetricsControllerClass) =>
      new MetricsControllerClass(
        this.config,
        this.dataModelerService,
        this.rillDeveloperService
      ).setup(this.app, "/api")
    );

    // TODO: This should be replaced by a better assignment of instance id once nodejs server is replaced completely bu runtime
    this.app.get("/api/v1/runtime/instance-id", (req: Request, res: Response) =>
      res.json({
        data: {
          instanceId: this.dataModelerService
            .getDatabaseService()
            .getDatabaseClient()
            .getInstanceId(),
        },
      })
    );

    // Temporary mirror to test before new CLI is merged
    this.app.get("/local/config", (req: Request, res: Response) =>
      res.json({
        instanceId: this.dataModelerService
          .getDatabaseService()
          .getDatabaseClient()
          .getInstanceId(),
        install_id: this.config.local.installId,
        project_id: this.dataModelerService
          .getStateService()
          .getApplicationState().projectId,
        is_dev: this.config.local.isDev,
      })
    );
  }
}
