import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Router, Request, Response } from "express";

export class MetricsDimensionController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    // metrics ID is here because express doesn't forward params from router path
    router.put("/:id/dimensions", (req: Request, res: Response) =>
      this.handleCreate(req, res)
    );
    router.post(
      "/:id/dimensions/:dimId/updateColumn",
      (req: Request, res: Response) => this.handleUpdateColumn(req, res)
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
