import { parse } from "yaml";
import { createRuntimeServiceGetFile } from "../../../runtime-client";

export function useDashboardPolicyCheck(instanceId: string, filePath: string) {
  return createRuntimeServiceGetFile(instanceId, filePath, {
    query: {
      select: (data) => {
        const yamlObj = parse(data?.blob);
        const securityPolicy = yamlObj?.security;
        return !!securityPolicy;
      },
    },
  });
}
