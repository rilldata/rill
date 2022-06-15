import { RillDeveloperController } from "$server/controllers/RillDeveloperController";
import type { Router, Request, Response } from "express";
import { RillRequestContext } from "$common/rill-developer-service/RillRequestContext";
import { RillActionsChannel } from "$common/utils/RillActionsChannel";

export class MetricsDefinitionController extends RillDeveloperController {
  protected setupRouter(router: Router) {
    router.put("/", (req: Request, res: Response) => {
      return this.handleMetricsDefinitionCreate(req, res);
    });
  }

  private async handleMetricsDefinitionCreate(req: Request, res: Response) {
    const context = new RillRequestContext(new RillActionsChannel());
    const promise = this.wrapHttpStream(
      res,
      context.actionsChannel.getActions()
    );
    await this.rillDeveloperService.dispatch(
      context,
      "createMetricsDefinition",
      []
    );
    return promise;
  }
}
