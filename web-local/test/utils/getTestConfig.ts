import { DatabaseConfig } from "@rilldata/web-local/common/config/DatabaseConfig";
import { RootConfig } from "@rilldata/web-local/common/config/RootConfig";
import { ServerConfig } from "@rilldata/web-local/common/config/ServerConfig";
import { StateConfig } from "@rilldata/web-local/common/config/StateConfig";

export function getTestConfig(
  projectFolder: string,
  {
    profileWithUpdate,
    socketPort,
    autoSync,
    serveStaticFile,
  }: {
    profileWithUpdate?: boolean;
    socketPort?: number;
    autoSync?: boolean;
    serveStaticFile?: boolean;
  } = {}
) {
  profileWithUpdate ??= true;
  socketPort ??= 8080;
  autoSync ??= true;
  serveStaticFile ??= false;

  return new RootConfig({
    database: new DatabaseConfig({
      databaseName: ":memory:",
      spawnRuntime: false,
    }),
    state: new StateConfig({ autoSync: autoSync, syncInterval: 50 }),
    server: new ServerConfig({ serverPort: socketPort, serveStaticFile }),
    projectFolder,
    profileWithUpdate,
  });
}
