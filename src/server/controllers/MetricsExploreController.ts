import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Router, Request, Response } from "express";
import { EntityController } from "$server/controllers/EntityController";

export class MetricsExploreController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    router.post("/metrics/:id/time-series", (req: Request, res: Response) =>
      this.handleGetTimeSeries(req, res)
    );
    router.get("/metrics/:id/time-range", (req: Request, res: Response) =>
      this.handleGetTimeRange(req, res)
    );
    router.post("/metrics/:id/leaderboards", (req: Request, res: Response) =>
      this.handleGetLeaderboards(req, res)
    );
    router.post("/metrics/:id/big-number", (req: Request, res: Response) =>
      this.bigNumber(req, res)
    );
  }

  private async handleGetTimeSeries(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(context, "generateTimeSeries", [
        req.params.id,
        req.body,
      ])
    );
  }

  private async handleGetTimeRange(req: Request, res: Response) {
    await EntityController.wrapAction(res, (context) =>
      this.rillDeveloperService.dispatch(context, "getTimeRange", [
        req.params.id,
      ])
    );
  }

  private async handleGetLeaderboards(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(context, "getLeaderboardValues", [
        req.params.id,
        req.body.measureId,
        req.body.filters,
      ])
    );
  }

  private async bigNumber(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(context, "getBigNumber", [
        req.params.id,
        req.body,
      ])
    );
  }
}
