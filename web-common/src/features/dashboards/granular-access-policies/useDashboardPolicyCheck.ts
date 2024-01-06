import { parse } from "yaml";
import { createRuntimeServiceGetFile } from "../../../runtime-client";
import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
import { EntityType } from "../../entity-management/types";

export function useDashboardPolicyCheck(
  instanceId: string,
  dashboardName: string,
) {
  return createRuntimeServiceGetFile(
    instanceId,
    getFilePathFromNameAndType(dashboardName, EntityType.MetricsDefinition),
    {
      query: {
        select: (data) => {
          const yamlObj = parse(data?.blob);
          const securityPolicy = yamlObj?.security;
          return !!securityPolicy;
        },
      },
    },
  );
}
