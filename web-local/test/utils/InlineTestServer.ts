import { TestServer } from "./TestServer";
import { RillDeveloper } from "$web-local/server/RillDeveloper";
import type { RootConfig } from "$web-local/common/config/RootConfig";
import type { RillDeveloperService } from "$web-local/common/rill-developer-service/RillDeveloperService";
import {
  expressServerFactory,
  rillDeveloperServiceFactory,
} from "$web-local/server/serverFactory";
import type { ExpressServer } from "$web-local/server/ExpressServer";

export class InlineTestServer extends TestServer {
  public readonly rillDeveloper: RillDeveloper;
  public readonly rillDeveloperService: RillDeveloperService;
  public readonly expressServer: ExpressServer;
  public readonly app: Express.Application;

  constructor(public readonly config: RootConfig) {
    const rillDeveloper = RillDeveloper.getRillDeveloper(config);
    super(
      rillDeveloper.dataModelerService,
      rillDeveloper.dataModelerStateService,
      rillDeveloper.dataModelerService.getDatabaseService()
    );

    this.rillDeveloper = rillDeveloper;
    this.rillDeveloperService = rillDeveloperServiceFactory(this.rillDeveloper);
    this.expressServer = expressServerFactory(
      config,
      this.rillDeveloper,
      this.rillDeveloperService
    );

    this.app = this.expressServer.app;
  }

  public async init() {
    await this.rillDeveloper.init();
    await this.expressServer.init();
  }

  public async destroy() {
    await this.rillDeveloper.destroy();
    await this.expressServer.destroy();
  }
}
