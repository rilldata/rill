import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Router } from "express";

export class TelemetryController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    router.post("/v1/telemetry/:eventType", async (req, res) => {
      await this.dataModelerService.metricsService.dispatch(
        req.params.eventType as any,
        [...req.body]
      );
      res.setHeader("Content-Type", "application/json");
      res.status(200);
      res.send({});
    });
  }
}
