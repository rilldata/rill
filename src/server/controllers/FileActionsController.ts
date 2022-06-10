import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Router, Request, Response } from "express";
import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";
import path from "path";

export class FileActionsController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    router.post("/table-upload", (req: any, res) => {
      this.handleFileUpload((req as any).files.file);
      res.send("OK");
    });
    router.get("/export", async (req, res) => {
      this.handleFileExport(req, res);
    });
  }

  private async handleFileUpload(file: {
    name: string;
    tempFilePath: string;
    mimetype: string;
    data: Buffer;
    size: number;
    mv: (string) => void;
  }) {
    const filePath = `${this.config.projectFolder}/tmp/${file.name}`;
    file.mv(filePath);
    await this.dataModelerService.dispatch("addOrUpdateTableFromFile", [
      filePath,
    ]);
  }

  private async handleFileExport(req: Request, res: Response) {
    const modelId = req.query.id as string;
    const exportType =
      req.query.type === "csv" ? "exportToCsv" : "exportToParquet";
    const fileName = decodeURIComponent(req.query.fileName as string);
    const exportResp = await this.dataModelerService.dispatch(exportType, [
      modelId,
      fileName,
    ]);
    if (exportResp.status === ActionStatus.Success) {
      res.setHeader("Content-Type", "application/octet-stream");
      res.setHeader(
        "Content-Disposition",
        `attachment; filename="${fileName}"`
      );
      res.sendFile(
        FileActionsController.getFile(
          `${this.config.database.exportFolder}/${fileName}`
        )
      );
    } else {
      res.status(500);
      res.send(
        `Failed to export.\n${exportResp.messages
          .map((message) => message.message)
          .join("\n")}`
      );
    }
  }

  private static getFile(filePath: string) {
    return path.isAbsolute(filePath)
      ? filePath
      : `${process.cwd()}/${filePath}`;
  }
}
