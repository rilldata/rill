import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Router, Request, Response } from "express";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export class MetricsMeasureController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    // metrics ID is here because express doesn't forward params from router path
    router.get("/:id/measures", (req: Request, res: Response) =>
      this.handleGetForMetricsDef(req, res)
    );
    router.put("/:id/measures", (req: Request, res: Response) =>
      this.handleCreate(req, res)
    );
    router.post("/:id/measures/:measureId", (req: Request, res: Response) =>
      this.handleUpdateMeasure(req, res)
    );
    router.post(
      "/:id/measures/:measureId/updateExpression",
      (req: Request, res: Response) => this.handleUpdateExpression(req, res)
    );
    router.post(
      "/:id/measures/:measureId/updateSqlName",
      (req: Request, res: Response) => this.handleUpdateSqlName(req, res)
    );
  }

  private async handleGetForMetricsDef(req: Request, res: Response) {
    if (!req.params.id || req.params.id === "undefined") return [];
    res.setHeader("ContentType", "application/json");
    res.send(
      JSON.stringify({
        data: this.rillDeveloperService.dataModelerStateService
          .getEntityStateService(
            EntityType.MetricsDefinition,
            StateType.Persistent
          )
          .getById(req.params.id).measures,
      })
    );
  }

  private async handleCreate(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(context, "addNewMeasure", [
        req.params.id,
      ])
    );
  }

  private async handleUpdateMeasure(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(context, "updateMeasure", [
        req.params.id,
        req.params.measureId,
        req.body,
      ])
    );
  }

  private async handleUpdateExpression(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(context, "updateMeasureExpression", [
        req.params.id,
        req.params.measureId,
        req.body.expression,
      ])
    );
  }

  private async handleUpdateSqlName(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(context, "updateMeasureSqlName", [
        req.params.id,
        req.params.measureId,
        req.body.sqlName,
      ])
    );
  }
}
