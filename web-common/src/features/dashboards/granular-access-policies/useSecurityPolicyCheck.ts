import { parse } from "yaml";
import type { RuntimeClient } from "../../../runtime-client/v2";
import { createRuntimeServiceGetFile } from "../../../runtime-client/v2/gen/runtime-service";

export function useDashboardPolicyCheck(
  client: RuntimeClient,
  filePath: string,
) {
  return createRuntimeServiceGetFile(
    client,
    {
      path: filePath,
    },
    {
      query: {
        select: (data) => {
          if (!data.blob) return false;
          const yamlObj = parse(data.blob);
          const securityPolicy = yamlObj?.security;
          return !!securityPolicy;
        },
      },
    },
  );
}

export function useRillYamlPolicyCheck(client: RuntimeClient) {
  return createRuntimeServiceGetFile(
    client,
    {
      path: "rill.yaml",
    },
    {
      query: {
        select: (data) => {
          if (!data.blob) return false;
          const yamlObj = parse(data.blob);
          const exploresSecurityPolicy = yamlObj?.explores?.security;
          const metricsViewsSecurityPolicy = yamlObj?.metricsViews?.security;
          return !!exploresSecurityPolicy || !!metricsViewsSecurityPolicy;
        },
      },
    },
  );
}
