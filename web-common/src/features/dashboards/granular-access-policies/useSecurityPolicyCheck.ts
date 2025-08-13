import { parse } from "yaml";
import { createRuntimeServiceGetFile } from "../../../runtime-client";

export function useDashboardPolicyCheck(instanceId: string, filePath: string) {
  return createRuntimeServiceGetFile(
    instanceId,
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

export function useRillYamlPolicyCheck(instanceId: string) {
  return createRuntimeServiceGetFile(
    instanceId,
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
