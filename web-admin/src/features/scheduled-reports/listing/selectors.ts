import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";

export function useReports(instanceId: string) {
  return createRuntimeServiceListResources(instanceId, {
    kind: ResourceKind.Report,
  });
}
