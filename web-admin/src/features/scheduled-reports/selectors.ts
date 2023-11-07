import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  createRuntimeServiceGetResource,
  createRuntimeServiceListResources,
} from "@rilldata/web-common/runtime-client";

export function useReports(instanceId: string) {
  return createRuntimeServiceListResources(instanceId, {
    kind: ResourceKind.Report,
  });
}

export function useReport(instanceId: string, name: string) {
  return createRuntimeServiceGetResource(instanceId, {
    "name.name": name,
    "name.kind": ResourceKind.Report,
  });
}

export function useReportDashboardName(instanceId: string, name: string) {
  return createRuntimeServiceGetResource(
    instanceId,
    {
      "name.name": name,
      "name.kind": ResourceKind.Report,
    },
    {
      query: {
        select: (data) => {
          const queryArgsJson = JSON.parse(
            data.resource.report.spec.queryArgsJson
          );

          return (
            queryArgsJson?.metrics_view_name ??
            queryArgsJson?.metricsViewName ??
            null
          );
        },
      },
    }
  );
}
