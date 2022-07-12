import "../moduleAlias";
import { RootConfig } from "$common/config/RootConfig";
import { RillDeveloper } from "./RillDeveloper";
import { ServerConfig } from "$common/config/ServerConfig";
import {
  expressServerFactory,
  rillDeveloperServiceFactory,
} from "$server/serverFactory";

const config = new RootConfig({
  server: new ServerConfig({ serveStaticFile: true }),
});
const rillDeveloper = RillDeveloper.getRillDeveloper(config);
const expressServer = expressServerFactory(
  config,
  rillDeveloper,
  rillDeveloperServiceFactory(rillDeveloper)
);
(async () => {
  await rillDeveloper.init();
  await expressServer.init();
})();
