import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Router, Request, Response } from "express";
import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";
import path from "path";
import { existsSync } from "fs";

export class FileActionsController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    router.post("/table-upload", (req: Request, res: Response) => {
      this.handleFileUpload((req as any).files.file, req.body.tableName);
      res.send("OK");
    });
    router.get("/export", async (req: Request, res: Response) =>
      this.handleFileExport(req, res)
    );
    router.get("/validate-table", async (req, res) =>
      this.handleTableValidation(req, res)
    );
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
    const fullPath = FileActionsController.getFile(
      `${this.config.database.exportFolder}/${fileName}`
    );
    if (existsSync(fullPath)) {
      res.setHeader("Content-Type", "application/octet-stream");
      res.setHeader(
        "Content-Disposition",
        `attachment; filename="${fileName}"`
      );
      res.sendFile(path.resolve(fullPath));
    } else {
      res.status(500);
      res.send(`Failed to export file ${fileName}`);
    }
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

  private static getFile(filePath: string) {
    return path.isAbsolute(filePath)
      ? filePath
      : `${process.cwd()}/${filePath}`;
  }
}
