import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Router, Request, Response } from "express";

export class MetricsMeasureController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    router.put("/:id/measures", (req: Request, res: Response) =>
      this.handleCreate(req, res)
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

  private async handleCreate(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(context, "addNewMeasure", [
        req.params.id,
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
