import { EntityController } from "$server/controllers/EntityController";
import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Request, Response, Router } from "express";

/**
 * Controller for metrics explore endpoints.
 * Based on rill runtime specs.
 */
export class MetricViewController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    router.get("/v1/metric-views/:id/meta", (req: Request, res: Response) =>
      this.handleGetMetricsMeta(req, res)
    );
    router.post(
      "/v1/metric-views/:id/timeseries",
      (req: Request, res: Response) => this.handleGetTimeSeries(req, res)
    );
    router.post(
      "/v1/metric-views/:id/toplist/:dimension",
      (req: Request, res: Response) => this.handleGetLeaderboards(req, res)
    );
    router.post("/v1/metric-views/:id/totals", (req: Request, res: Response) =>
      this.bigNumber(req, res)
    );
  }

  private async handleGetMetricsMeta(req: Request, res: Response) {
    await EntityController.wrapAction(res, (context) =>
      this.rillDeveloperService.dispatch(context, "getMetricViewMeta", [
        req.params.id,
      ])
    );
  }

  private async handleGetTimeSeries(req: Request, res: Response) {
    return EntityController.wrapAction(res, (context) =>
      this.rillDeveloperService.dispatch(context, "getMetricViewTimeSeries", [
        req.params.id,
        req.body,
      ])
    );
  }

  private async handleGetLeaderboards(req: Request, res: Response) {
    return EntityController.wrapAction(res, (context) =>
      this.rillDeveloperService.dispatch(context, "getMetricViewTopList", [
        req.params.id,
        req.params.dimension,
        req.body,
      ])
    );
  }

  private async bigNumber(req: Request, res: Response) {
    return EntityController.wrapAction(res, (context) =>
      this.rillDeveloperService.dispatch(context, "getMetricViewTotals", [
        req.params.id,
        req.body,
      ])
    );
  }
}
