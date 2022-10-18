import "../moduleAlias";
import { RootConfig } from "@rilldata/web-local/common/config/RootConfig";
import path from "node:path";
import { RillDeveloper } from "./RillDeveloper";
import { ServerConfig } from "@rilldata/web-local/common/config/ServerConfig";
import {
  expressServerFactory,
  rillDeveloperServiceFactory,
} from "./serverFactory";
import { LocalConfig } from "@rilldata/web-local/common/config/LocalConfig";

let ProjectFolder: string;
// use `RILL_PROJECT` to override project folder while running in dev mode.
// this can be helpful when testing fresh projects without needing to delete existing one.
if (process.env.RILL_PROJECT) {
  ProjectFolder = path.isAbsolute(process.env.RILL_PROJECT)
    ? process.env.RILL_PROJECT
    : path.resolve("../" + process.env.RILL_PROJECT);
} else {
  ProjectFolder = path.resolve("..");
}

const config = new RootConfig({
  projectFolder: ProjectFolder,
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
