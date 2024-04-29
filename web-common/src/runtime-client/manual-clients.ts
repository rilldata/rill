// This files contains clients that are not written through GRPC

import { V1ComponentSpecResolverProperties } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import httpClient from "@rilldata/web-common/runtime-client/http-client";
import { QueryClient } from "@tanstack/query-core";
import { createQuery } from "@tanstack/svelte-query";

export type V1RuntimeGetConfig = {
  instance_id: string;
  grpc_port: number;
  install_id: string;
  project_path: string;
  user_id: string;
  version: string;
  build_commit: string;
  is_dev: boolean;
  analytics_enabled: boolean;
  readonly: boolean;
};
export const runtimeServiceGetConfig =
  async (): Promise<V1RuntimeGetConfig> => {
    return httpClient({
      url: "/local/config",
      method: "GET",
    });
  };

export const runtimeServiceFileUpload = async (
  instanceId: string,
  filePath: string,
  formData: FormData,
) => {
  return httpClient({
    url: `/v1/instances/${instanceId}/files/upload/-/${filePath}`,
    method: "POST",
    data: formData,
    headers: {},
  });
};

export function runtimeServiceGetChartData(
  instanceId: string,
  chartName: string,
  signal?: AbortSignal,
) {
  return httpClient({
    url: `/v1/instances/${instanceId}/components/${chartName}/data`,
    method: "GET",
    headers: {},
    signal,
  });
}

export function createRuntimeServiceGetChartData(
  queryClient: QueryClient,
  instanceId: string,
  chartName: string,
  // we need this till we figure out why last updated is not changing on charts
  props: V1ComponentSpecResolverProperties | undefined,
) {
  return createQuery(
    [`/v1/instances/${instanceId}/components/${chartName}/data`, props],
    {
      queryFn: ({ signal }) =>
        runtimeServiceGetChartData(instanceId, chartName, signal),
      enabled: !!instanceId && !!chartName,
      queryClient,
    },
  );
}
