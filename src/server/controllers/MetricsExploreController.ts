import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Router, Request, Response } from "express";
import { RillRequestContext } from "$common/rill-developer-service/RillRequestContext";
import { RillActionsChannel } from "$common/utils/RillActionsChannel";

export class MetricsExploreController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    router.post("/metrics/:id/leaderboards", (req: Request, res: Response) =>
      this.handleGetLeaderboards(req, res)
    );
    router.post("/metrics/:id/bigNumber", (req: Request, res: Response) =>
      this.bigNumber(req, res)
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
    res.setHeader("ContentType", "application/json");
    const context = new RillRequestContext(new RillActionsChannel());
    res.send(
      JSON.stringify({
        data: await this.rillDeveloperService.dispatch(
          context,
          "getBigNumber",
          [req.params.id, req.body.measureId, req.body.filters]
        ),
      })
    );
  }
}
