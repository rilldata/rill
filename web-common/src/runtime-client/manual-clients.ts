// This files contains clients that are not written through GRPC

import httpClient from "@rilldata/web-common/runtime-client/http-client";

export type V1RuntimeGetConfig = {
  instance_id: string;
  grpc_port: number;
  install_id: string;
  project_path: string;
  version: string;
  build_commit: string;
  is_dev: boolean;
  analytics_enabled: boolean;
  readonly: boolean;
};

export let globalConfig: V1RuntimeGetConfig;

export const runtimeServiceGetConfig =
  async (): Promise<V1RuntimeGetConfig> => {
    if (!globalConfig) {
      globalConfig = await httpClient({
        url: "/local/config",
        method: "GET",
      });
    }
    return globalConfig;
  };

export const runtimeServiceFileUpload = async (
  instanceId: string,
  filePath: string,
  formData: FormData
) => {
  return httpClient({
    url: `/v1/instances/${instanceId}/files/upload/-/${filePath}`,
    method: "POST",
    data: formData,
    headers: {},
  });
};
