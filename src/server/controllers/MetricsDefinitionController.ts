import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Request, Response, Router } from "express";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

export class MetricsDefinitionController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    router.get("/", (req: Request, res: Response) =>
      this.handleGetAll(req, res)
    );
    router.put("/", (req: Request, res: Response) =>
      this.handleCreate(req, res)
    );
    router.post("/:id/updateModel", (req: Request, res: Response) =>
      this.handleModelUpdate(req, res)
    );
    router.post("/:id/updateTimestamp", (req: Request, res: Response) =>
      this.handleTimestampUpdate(req, res)
    );
    router.delete("/:id", (req: Request, res: Response) =>
      this.handleDelete(req, res)
    );
  }

  private async handleGetAll(req: Request, res: Response) {
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

  private async handleCreate(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(context, "createMetricsDefinition", [])
    );
  }

  private async handleModelUpdate(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(
        context,
        "updateMetricsDefinitionModel",
        [req.params.id, req.body.modelId]
      )
    );
  }

  private async handleTimestampUpdate(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(
        context,
        "updateMetricsDefinitionTimestamp",
        [req.params.id, req.body.timeDimension]
      )
    );
  }

  private async handleDelete(req: Request, res: Response) {
    return this.wrapHttpStream(res, (context) =>
      this.rillDeveloperService.dispatch(context, "deleteMetricsDefinition", [
        req.params.id,
      ])
    );
  }
}
