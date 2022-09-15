import "../moduleAlias";
import { RootConfig } from "$common/config/RootConfig";
import { RillDeveloper } from "./RillDeveloper";
import { ServerConfig } from "$common/config/ServerConfig";
import {
  expressServerFactory,
  rillDeveloperServiceFactory,
} from "$server/serverFactory";
import { LocalConfig } from "$common/config/LocalConfig";

const config = new RootConfig({
  // use `RILL_PROJECT` to override project folder while running in dev mode.
  // this can be helpful when testing fresh projects without needing to delete existing one.
  projectFolder: process.env.RILL_PROJECT ?? ".",
  server: new ServerConfig({ serveStaticFile: true }),
  local: new LocalConfig({ isDev: true }),
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
