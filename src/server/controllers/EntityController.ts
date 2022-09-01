import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Request, Response, Router } from "express";
import type { ActionResponse } from "$common/data-modeler-service/response/ActionResponse";
import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";
import { ActionResponseFactory } from "$common/data-modeler-service/response/ActionResponseFactory";
import { RillRequestContext } from "$common/rill-developer-service/RillRequestContext";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { ActionDefinitionError } from "$common/errors/ActionDefinitionError";

export abstract class EntityController extends RillDeveloperController {
  protected static entityPath: string;
  protected static entityType: EntityType;

  protected setupRouter(router: Router) {
    router.get(
      `/${this.getClass().entityPath}`,
      (req: Request, res: Response) => this.getAll(req, res)
    );
    router.get(
      `/${this.getClass().entityPath}/:id`,
      (req: Request, res: Response) => this.getOne(req, res)
    );
    router.put(
      `/${this.getClass().entityPath}`,
      (req: Request, res: Response) => this.create(req, res)
    );
    router.post(
      `/${this.getClass().entityPath}/:id`,
      (req: Request, res: Response) => this.update(req, res)
    );
    router.delete(
      `/${this.getClass().entityPath}/:id`,
      (req: Request, res: Response) => this.delete(req, res)
    );
  }

  protected abstract getAll(req: Request, res: Response): Promise<void>;

  protected async getOne(req: Request, res: Response): Promise<void> {
    res.setHeader("ContentType", "application/json");
    res.send(
      JSON.stringify({
        data: this.rillDeveloperService.dataModelerStateService
          .getEntityStateService(
            this.getClass().entityType,
            StateType.Persistent
          )
          .getById(req.params.id),
      })
    );
  }

  protected async create(req: Request, res: Response): Promise<void> {
    res.setHeader("ContentType", "application/json");
    await EntityController.wrapAction(res, (context) =>
      this.createAction(context, req)
    );
  }
  protected abstract createAction(
    context: RillRequestContext,
    req: Request
  ): Promise<ActionResponse>;

  protected async update(req: Request, res: Response): Promise<void> {
    res.setHeader("ContentType", "application/json");
    await EntityController.wrapAction(res, (context) =>
      this.updateAction(context, req)
    );
  }
  protected abstract updateAction(
    context: RillRequestContext,
    req: Request
  ): Promise<ActionResponse>;

  protected async delete(req: Request, res: Response): Promise<void> {
    res.setHeader("ContentType", "application/json");
    await EntityController.wrapAction(res, (context) =>
      this.deleteAction(context, req)
    );
  }
  protected abstract deleteAction(
    context: RillRequestContext,
    req: Request
  ): Promise<ActionResponse>;

  private getClass(): typeof EntityController {
    return this.constructor as typeof EntityController;
  }

  public static async wrapAction(
    res: Response,
    callback: (context: RillRequestContext) => Promise<ActionResponse>
  ) {
    res.setHeader("Content-Type", "application/json");
    try {
      const response = await callback(RillRequestContext.getNewContext());
      if (!response || response.status === ActionStatus.Failure) {
        res.status(500);
        res.send(
          JSON.stringify(
            response ??
              ActionResponseFactory.getErrorResponse(
                new ActionDefinitionError("Missing response")
              )
          )
        );
      } else {
        res.status(200);
        res.json({
          ...response,
        });
      }
    } catch (err) {
      res.status(500);
      res.send(JSON.stringify(ActionResponseFactory.getErrorResponse(err)));
    }
  }
}
