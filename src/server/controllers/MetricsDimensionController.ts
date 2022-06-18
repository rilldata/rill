import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Router, Request, Response } from "express";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export class MetricsDimensionController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    // metrics ID is here because express doesn't forward params from router path
    router.get("/:id/dimensions", (req: Request, res: Response) =>
      this.handleGetForMetricsDef(req, res)
    );
    router.put("/:id/dimensions", (req: Request, res: Response) =>
      this.handleCreate(req, res)
    );
    router.post("/:id/dimensions/:dimId", (req: Request, res: Response) =>
      this.handleUpdateDimension(req, res)
    );
    router.post(
      "/:id/dimensions/:dimId/updateColumn",
      (req: Request, res: Response) => this.handleUpdateColumn(req, res)
    );
  }

  private async handleGetForMetricsDef(req: Request, res: Response) {
    res.setHeader("ContentType", "application/json");
    res.send(
      JSON.stringify({
        data: this.rillDeveloperService.dataModelerStateService
          .getEntityStateService(
            EntityType.MetricsDefinition,
            StateType.Persistent
          )
          .getById(req.params.id).dimensions,
      })
    );
  }

  private async handleCreate(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(context, "addNewDimension", [
        req.params.id,
        req.body.column,
      ])
    );
  }

  private async handleUpdateDimension(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(context, "updateDimension", [
        req.params.id,
        req.params.dimId,
        req.body,
      ])
    );
  }

  private async handleUpdateColumn(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(context, "updateDimensionColumn", [
        req.params.id,
        req.params.dimId,
        req.body.column,
      ])
    );
  }
}
