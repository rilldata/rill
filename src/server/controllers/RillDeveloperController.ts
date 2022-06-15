import type { Application } from "express";
import { Response, Router } from "express";
import type { RootConfig } from "$common/config/RootConfig";
import type { RillDeveloperService } from "$common/rill-developer-service/RillDeveloperService";

export abstract class RillDeveloperController {
  public constructor(
    protected readonly config: RootConfig,
    protected readonly rillDeveloperService: RillDeveloperService
  ) {}

  public setup(app: Application, path: string) {
    const router = Router();
    this.setupRouter(router);
    app.use(path, router);
  }

  protected abstract setupRouter(router: Router);

  protected async wrapHttpStream(res: Response, generator: AsyncGenerator) {
    res.writeHead(200, {
      Connection: "keep-alive",
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
    });
    for await (const data of generator) {
      res.write(`data: ${JSON.stringify(data)}`);
    }
    res.write("\n\n");
  }
}
