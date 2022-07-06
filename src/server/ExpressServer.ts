import express, { Request, Response } from "express";
import http from "http";
import cors from "cors";
import fileUpload from "express-fileupload";
import type { RootConfig } from "$common/config/RootConfig";
import { SocketServer } from "$server/SocketServer";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type { SocketNotificationService } from "$common/socket/SocketNotificationService";
import type { MetricsService } from "$common/metrics-service/MetricsService";
import { existsSync, mkdirSync } from "fs";
import path from "node:path";
import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";

const STATIC_FILES = `${__dirname}/../../build`;

export class ExpressServer {
  private readonly app: express.Application;
  private readonly server: http.Server;
  private readonly socketServer: SocketServer;

  constructor(
    private readonly config: RootConfig,
    private readonly dataModelerService: DataModelerService,
    dataModelerStateService: DataModelerStateService,
    notificationService: SocketNotificationService,
    metricsService: MetricsService
  ) {
    this.app = express();
    this.server = http.createServer(this.app);

    this.app.use(
      cors({
        origin: this.config.server.uiUrl,
      })
    );

    const tmpFolder = `${config.projectFolder}/tmp`;
    if (!existsSync(tmpFolder)) mkdirSync(tmpFolder);
    this.app.use(
      fileUpload({
        useTempFiles: true,
        tempFileDir: tmpFolder,
      })
    );

    this.app.post("/api/table-upload", (req: any, res) => {
      if (req.body?.tableName) {
        this.handleFileUpload(
          (req as any).files.file,
          (req as any).body.tableName
        );
      } else {
        this.handleFileUpload((req as any).files.file);
      }
      res.send("OK");
    });
    this.app.get("/api/export", async (req, res) => {
      this.handleFileExport(req, res);
    });
    this.app.get("/api/validate-table", async (req, res) => {
      this.handleTableValidation(req, res);
    });

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

  private async handleTableValidation(req: Request, res: Response) {
    const tableName = decodeURIComponent(req.query.tableName as string);

    const response = await this.dataModelerService.dispatch(
      "validateTableName",
      [tableName]
    );

    if (response.status === ActionStatus.Success) {
      if (!response.messages.length) {
        res.json({
          isDuplicate: false,
        });
      } else {
        res.json({
          isDuplicate: true,
          name: response.messages[0].message,
        });
      }
    } else {
      res.status(500);
      res.send(`Failed to validate table name ${tableName}`);
    }
  }

  private async handleFileUpload(
    file: {
      name: string;
      tempFilePath: string;
      mimetype: string;
      data: Buffer;
      size: number;
      mv: (string) => void;
    },
    tableName?: string
  ) {
    const filePath = `${this.config.projectFolder}/tmp/${file.name}`;
    file.mv(filePath);

    if (tableName) {
      await this.dataModelerService.dispatch("addOrUpdateTableFromFile", [
        filePath,
        tableName,
      ]);
    } else {
      await this.dataModelerService.dispatch("addOrUpdateTableFromFile", [
        filePath,
      ]);
    }
  }

  private async handleFileExport(req: Request, res: Response) {
    const fileName = decodeURIComponent(req.query.fileName as string);
    const fullPath = ExpressServer.getAbsoluteFilePath(
      `${this.config.database.exportFolder}/${fileName}`
    );
    if (existsSync(fullPath)) {
      res.setHeader("Content-Type", "application/octet-stream");
      res.setHeader(
        "Content-Disposition",
        `attachment; filename="${fileName}"`
      );
      res.sendFile(fullPath);
    } else {
      res.status(500);
      res.send(`Failed to export file ${fileName}`);
    }
  }

  private static getAbsoluteFilePath(filePath: string) {
    return path.isAbsolute(filePath)
      ? filePath
      : `${process.cwd()}/${filePath}`;
  }
}
