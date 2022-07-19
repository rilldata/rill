import type { Application } from "express";
import { Response, Router } from "express";
import type { RootConfig } from "$common/config/RootConfig";
import type { RillDeveloperService } from "$common/rill-developer-service/RillDeveloperService";
import { RillRequestContext } from "$common/rill-developer-service/RillRequestContext";
import type { ActionResponse } from "$common/data-modeler-service/response/ActionResponse";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export abstract class RillDeveloperController {
  public constructor(
    protected readonly config: RootConfig,
    protected readonly dataModelerService: DataModelerService,
    protected readonly rillDeveloperService: RillDeveloperService
  ) {}

  public setup(app: Application, path: string) {
    const router = Router();
    this.setupRouter(router);
    app.use(path, router);
  }

  protected abstract setupRouter(router: Router);

  protected async wrapHttpStream(
    res: Response,
    callback: (
      context: RillRequestContext<EntityType, StateType>
    ) => Promise<ActionResponse>
  ) {
    const context = RillRequestContext.getNewContext();
    res.writeHead(200, {
      Connection: "keep-alive",
      "Content-Type": "application/json",
      "Cache-Control": "no-cache",
    });
    const promise = callback(context);
    for await (const data of context.actionsChannel.getActions()) {
      res.write(JSON.stringify(data) + "\n");
    }
    await promise;
    res.end();
  }
}
