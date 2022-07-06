import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Router, Request, Response } from "express";
import path from "path";
import { existsSync } from "node:fs";

export class FileActionsController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    router.post("/table-upload", (req: Request, res: Response) => {
      this.handleFileUpload((req as any).files.file);
      res.send("OK");
    });
    router.get("/export", async (req: Request, res: Response) => {
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
      res.sendFile(fullPath);
    } else {
      res.status(500);
      res.send(`Failed to export file ${fileName}`);
    }
  }

  private static getFile(filePath: string) {
    return path.isAbsolute(filePath)
      ? filePath
      : `${process.cwd()}/${filePath}`;
  }
}
