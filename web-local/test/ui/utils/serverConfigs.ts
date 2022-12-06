import { useTestBrowser, useTestServer } from "./useTestServer";

export const ServerConfigs: {
  [k in string]: {
    port: number;
    projectFolder: string;
  };
} = {
  sources: {
    port: 8081,
    projectFolder: "temp/sources",
  },
  models: {
    port: 8082,
    projectFolder: "temp/models",
  },
  dashboards: {
    port: 8083,
    projectFolder: "temp/dashboards",
  },
};

export function useRegisteredServer(name: string) {
  const config = ServerConfigs[name];
  useTestServer(config.port, config.projectFolder);
  return useTestBrowser(config.port);
}
