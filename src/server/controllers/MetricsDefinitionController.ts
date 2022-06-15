import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Router, Request, Response } from "express";

export class MetricsDefinitionController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    router.put("/", (req: Request, res: Response) =>
      this.handleCreate(req, res)
    );
    router.post("/:id/updateModel", (req: Request, res: Response) =>
      this.handleModelUpdate(req, res)
    );
    router.post("/:id/updateTimestamp", (req: Request, res: Response) =>
      this.handleTimestampUpdate(req, res)
    );
  }

  private async handleCreate(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(context, "createMetricsDefinition", [])
    );
  }

  private async handleModelUpdate(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(
        context,
        "updateMetricsDefinitionModel",
        [req.params.id, req.body.modelId]
      )
    );
  }

  private async handleTimestampUpdate(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(
        context,
        "updateMetricsDefinitionTimestamp",
        [req.params.id, req.body.timeDimension]
      )
    );
  }
}
