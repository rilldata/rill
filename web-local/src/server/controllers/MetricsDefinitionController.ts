import type { Request, Response, Router } from "express";
import {
  EntityType,
  StateType,
} from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { EntityController } from "./EntityController";
import type { RillRequestContext } from "@rilldata/web-local/common/rill-developer-service/RillRequestContext";
import type { ActionResponse } from "@rilldata/web-local/common/data-modeler-service/response/ActionResponse";

export class MetricsDefinitionController extends EntityController {
  protected static entityPath = "metrics";
  protected static entityType = EntityType.MetricsDefinition;

  protected setupRouter(router: Router) {
    super.setupRouter(router);
    router.post(
      "/metrics/:id/generate-measures-dimensions",
      (req: Request, res: Response) =>
        this.handleGenerateMeasuresAndDimensions(req, res)
    );
  }

  protected async getAll(req: Request, res: Response) {
    res.setHeader("ContentType", "application/json");
    res.send(
      JSON.stringify({
        data: this.rillDeveloperService.dataModelerStateService
          .getEntityStateService(
            EntityType.MetricsDefinition,
            StateType.Persistent
          )
          .getCurrentState().entities,
      })
    );
  }

  protected createAction(
    context: RillRequestContext,
    req: Request
  ): Promise<ActionResponse> {
    return this.rillDeveloperService.dispatch(
      context,
      "createMetricsDefinition",
      [req.body]
    );
  }

  protected updateAction(
    context: RillRequestContext,
    req: Request
  ): Promise<ActionResponse> {
    return this.rillDeveloperService.dispatch(
      context,
      "updateMetricsDefinition",
      [req.params.id, req.body]
    );
  }

  protected deleteAction(
    context: RillRequestContext,
    req: Request
  ): Promise<ActionResponse> {
    return this.rillDeveloperService.dispatch(
      context,
      "deleteMetricsDefinition",
      [req.params.id]
    );
  }

  protected async handleGenerateMeasuresAndDimensions(
    req: Request,
    res: Response
  ) {
    await this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(
        context,
        "generateMeasuresAndDimensions",
        [req.params.id]
      )
    );
  }
}
