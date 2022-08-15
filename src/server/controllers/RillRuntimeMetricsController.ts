import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Request, Response, Router } from "express";
import { EntityController } from "$server/controllers/EntityController";

/**
 * Controller for metrics explore endpoints.
 * Based on rill runtime specs.
 */
export class RillRuntimeMetricsController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    router.get("/metric-views/:id/meta", (req: Request, res: Response) =>
      this.handleGetMetricsMeta(req, res)
    );
    router.post("/metric-views/:id/timeseries", (req: Request, res: Response) =>
      this.handleGetTimeSeries(req, res)
    );
    router.post(
      "/metric-views/:id/toplist/:dimension",
      (req: Request, res: Response) => this.handleGetLeaderboards(req, res)
    );
    router.post("/metric-views/:id/big-number", (req: Request, res: Response) =>
      this.bigNumber(req, res)
    );
  }

  private async handleGetTimeSeries(req: Request, res: Response) {
    return EntityController.wrapAction(res, (context) =>
      this.rillDeveloperService.dispatch(context, "getRuntimeTimeSeries", [
        req.params.id,
        req.body,
      ])
    );
  }

  private async handleGetMetricsMeta(req: Request, res: Response) {
    await EntityController.wrapAction(res, (context) =>
      this.rillDeveloperService.dispatch(context, "getRuntimeMetricsMeta", [
        req.params.id,
      ])
    );
  }

  private async handleGetLeaderboards(req: Request, res: Response) {
    return EntityController.wrapAction(res, (context) =>
      this.rillDeveloperService.dispatch(context, "getRuntimeTopList", [
        req.params.id,
        req.params.dimension,
        req.body,
      ])
    );
  }

  private async bigNumber(req: Request, res: Response) {
    return EntityController.wrapAction(res, (context) =>
      this.rillDeveloperService.dispatch(context, "getRuntimeBigNumber", [
        req.params.id,
        req.body,
      ])
    );
  }
}
