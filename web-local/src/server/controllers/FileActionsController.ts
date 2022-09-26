import { RillDeveloperController } from "./RillDeveloperController";
import type { Request, Response, Router } from "express";
import { ActionStatus } from "$web-local/common/data-modeler-service/response/ActionResponse";
import path from "path";
import { existsSync } from "fs";
import { EntityController } from "./EntityController";

interface FileUploadEntry {
  name: string;
  tempFilePath: string;
  mimetype: string;
  data: Buffer;
  size: number;
  mv: (string) => void;
}
type FileUploadRequest = Request & { files: { file: FileUploadEntry } };

export class FileActionsController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    router.post("/table-upload", (req: FileUploadRequest, res: Response) =>
      this.handleFileUpload(req, res)
    );
    router.get("/export", async (req: Request, res: Response) =>
      this.handleFileExport(req, res)
    );
    router.get("/validate-table", async (req, res) =>
      this.handleTableValidation(req, res)
    );
  }

  private async handleFileUpload(req: FileUploadRequest, res: Response) {
    if (!req.files?.file) {
      res.status(500);
      res.send(`Failed to import source`);
      return;
    }
    const filePath = `${this.config.projectFolder}/tmp/${req.files.file.name}`;
    req.files.file.mv(filePath);

    await EntityController.wrapAction(res, () => {
      if (req.body.tableName) {
        return this.dataModelerService.dispatch("addOrUpdateTableFromFile", [
          filePath,
          req.body.tableName,
          { shouldNotProfile: true },
        ]);
      } else {
        return this.dataModelerService.dispatch("addOrUpdateTableFromFile", [
          filePath,
          undefined,
          { shouldNotProfile: true },
        ]);
      }
    });
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
          data: {
            isDuplicate: false,
          },
        });
      } else {
        res.json({
          data: {
            isDuplicate: true,
            name: response.messages[0].message,
          },
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
