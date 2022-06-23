import type { Request, Response } from "express";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { EntityController } from "$server/controllers/EntityController";
import type { ActionResponse } from "$common/data-modeler-service/response/ActionResponse";
import type { RillRequestContext } from "$common/rill-developer-service/RillRequestContext";

export class MetricsDimensionController extends EntityController {
  protected static entityPath = "dimensions";
  protected static entityType = EntityType.DimensionDefinition;

  protected async getAll(req: Request, res: Response): Promise<void> {
    const metricsDefId = req.query.metricsDefId as string;
    const dimensionsStateService =
      this.rillDeveloperService.dataModelerStateService.getEntityStateService(
        EntityType.DimensionDefinition,
        StateType.Persistent
      );
    const dimensions = dimensionsStateService
      .getCurrentState()
      .entities.filter((dimension) => dimension.metricsDefId === metricsDefId);

    res.setHeader("ContentType", "application/json");
    res.send(
      JSON.stringify({
        data: dimensions,
      })
    );
  }

  protected createAction(
    context: RillRequestContext,
    req: Request
  ): Promise<ActionResponse> {
    return this.rillDeveloperService.dispatch(context, "addNewDimension", [
      req.body.metricsDefId,
      req.body.columnName,
    ]);
  }

  protected updateAction(
    context: RillRequestContext,
    req: Request
  ): Promise<ActionResponse> {
    return this.rillDeveloperService.dispatch(context, "updateDimension", [
      req.params.id,
      req.body,
    ]);
  }

  protected deleteAction(
    context: RillRequestContext,
    req: Request
  ): Promise<ActionResponse> {
    return this.rillDeveloperService.dispatch(context, "deleteDimension", [
      req.params.id,
    ]);
  }
}
