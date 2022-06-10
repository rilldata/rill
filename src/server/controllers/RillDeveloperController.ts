import type { Application } from "express";
import { Router } from "express";
import type { RootConfig } from "$common/config/RootConfig";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";

export abstract class RillDeveloperController {
  public constructor(
    protected readonly config: RootConfig,
    protected readonly dataModelerService: DataModelerService
  ) {}

  public setup(app: Application, path: string) {
    const router = Router();
    this.setupRouter(router);
    app.use(path, router);
  }

  protected abstract setupRouter(router: Router);
}
