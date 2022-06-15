import "../moduleAlias";
import { RootConfig } from "$common/config/RootConfig";
import { RillDeveloper } from "./RillDeveloper";
import type { SocketNotificationService } from "$common/socket/SocketNotificationService";
import { ExpressServer } from "./ExpressServer";
import { ServerConfig } from "$common/config/ServerConfig";
import { rillDeveloperServiceFactory } from "$server/serverFactory";

const config = new RootConfig({
  server: new ServerConfig({ serveStaticFile: true }),
});
const rillDeveloper = RillDeveloper.getRillDeveloper(config);
const expressServer = new ExpressServer(
  config,
  rillDeveloper.dataModelerService,
  rillDeveloperServiceFactory(rillDeveloper),
  rillDeveloper.dataModelerStateService,
  rillDeveloper.notificationService as SocketNotificationService,
  rillDeveloper.metricsService
);
(async () => {
  await rillDeveloper.init();
  await expressServer.init();
})();
