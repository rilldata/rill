import { DatabaseConfig } from "$common/config/DatabaseConfig";
import { RootConfig } from "$common/config/RootConfig";
import { ServerConfig } from "$common/config/ServerConfig";
import { StateConfig } from "$common/config/StateConfig";

export function getTestConfig(
  projectFolder: string,
  {
    profileWithUpdate,
    socketPort,
    autoSync,
  }: {
    profileWithUpdate?: boolean;
    socketPort?: number;
    autoSync?: boolean;
  } = {}
) {
  profileWithUpdate ??= true;
  socketPort ??= 8080;
  autoSync ??= true;

  return new RootConfig({
    database: new DatabaseConfig({
      databaseName: ":memory:",
      spawnRuntime: false,
    }),
    state: new StateConfig({ autoSync: autoSync, syncInterval: 50 }),
    server: new ServerConfig({ serverPort: socketPort }),
    projectFolder,
    profileWithUpdate,
  });
}
