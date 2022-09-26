import type { ActionResponse } from "@rilldata/web-local/common/data-modeler-service/response/ActionResponse";
import {
  EntityType,
  StateType,
} from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { RillRequestContext } from "@rilldata/web-local/common/rill-developer-service/RillRequestContext";
import { EntityController } from "./EntityController";
import type { Request, Response, Router } from "express";

export class MetricsMeasureController extends EntityController {
  protected static entityPath = "measures";
  protected static entityType = EntityType.MeasureDefinition;

  protected setupRouter(router: Router) {
    router.post(
      "/measures/validate-expression",
      (req: Request, res: Response) => this.validateExpression(req, res)
    );
    super.setupRouter(router);
  }

  protected async getAll(req: Request, res: Response): Promise<void> {
    const metricsDefId = req.query.metricsDefId as string;
    const measuresStateService =
      this.rillDeveloperService.dataModelerStateService.getEntityStateService(
        EntityType.MeasureDefinition,
        StateType.Persistent
      );
    const measures = measuresStateService
      .getCurrentState()
      .entities.filter((measure) => measure.metricsDefId === metricsDefId);

    res.setHeader("ContentType", "application/json");
    res.send(
      JSON.stringify({
        data: measures,
      })
    );
  }

  protected createAction(
    context: RillRequestContext,
    req: Request
  ): Promise<ActionResponse> {
    return this.rillDeveloperService.dispatch(context, "addNewMeasure", [
      req.body.metricsDefId,
    ]);
  }

  protected updateAction(
    context: RillRequestContext,
    req: Request
  ): Promise<ActionResponse> {
    return this.rillDeveloperService.dispatch(context, "updateMeasure", [
      req.params.id,
      req.body,
    ]);
  }

  protected deleteAction(
    context: RillRequestContext,
    req: Request
  ): Promise<ActionResponse> {
    return this.rillDeveloperService.dispatch(context, "deleteMeasure", [
      req.params.id,
    ]);
  }

  private async validateExpression(req: Request, res: Response) {
    await EntityController.wrapAction(res, async (context) =>
      this.rillDeveloperService.dispatch(context, "validateMeasureExpression", [
        req.body.metricsDefId,
        req.body.measureId,
        req.body.expression,
      ])
    );
  }
}
