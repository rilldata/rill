import { EntityController } from "$server/controllers/EntityController";
import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Request, Response, Router } from "express";

/**
 * Controller for metrics explore endpoints.
 * Based on rill runtime specs.
 */
export class MetricsViewController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    router.get("/v1/metrics-views/:id/meta", (req: Request, res: Response) =>
      this.handleGetMetricsMeta(req, res)
    );
    router.post(
      "/v1/metrics-views/:id/timeseries",
      (req: Request, res: Response) => this.handleGetTimeSeries(req, res)
    );
    router.post(
      "/v1/metrics-views/:id/toplist/:dimension",
      (req: Request, res: Response) => this.handleGetLeaderboards(req, res)
    );
    router.post("/v1/metrics-views/:id/totals", (req: Request, res: Response) =>
      this.handleGetTotals(req, res)
    );
  }

  private async handleGetMetricsMeta(req: Request, res: Response) {
    await EntityController.wrapAction(res, (context) =>
      this.rillDeveloperService.dispatch(context, "getMetricsViewMeta", [
        req.params.id,
      ])
    );
  }

  private async handleGetTimeSeries(req: Request, res: Response) {
    return EntityController.wrapAction(res, (context) =>
      this.rillDeveloperService.dispatch(context, "getMetricsViewTimeSeries", [
        req.params.id,
        req.body,
      ])
    );
  }

  private async handleGetLeaderboards(req: Request, res: Response) {
    return EntityController.wrapAction(res, (context) =>
      this.rillDeveloperService.dispatch(context, "getMetricsViewTopList", [
        req.params.id,
        req.params.dimension,
        req.body,
      ])
    );
  }

  private async handleGetTotals(req: Request, res: Response) {
    return EntityController.wrapAction(res, (context) =>
      this.rillDeveloperService.dispatch(context, "getMetricsViewTotals", [
        req.params.id,
        req.body,
      ])
    );
  }
}
