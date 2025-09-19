import { createAdminServiceSearchProjectUsers } from "@rilldata/web-admin/client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { getDashboardNameFromReport } from "@rilldata/web-common/features/scheduled-reports/utils";
import {
  createRuntimeServiceGetResource,
  createRuntimeServiceListResources,
} from "@rilldata/web-common/runtime-client";
import { createSmartRefetchInterval } from "@rilldata/web-admin/lib/refetch-interval-store";

export function useReports(instanceId: string, enabled = true) {
  return createRuntimeServiceListResources(
    instanceId,
    {
      kind: ResourceKind.Report,
    },
    {
      query: {
        enabled: enabled && !!instanceId,
        refetchOnMount: true,
        refetchInterval: createSmartRefetchInterval,
      },
    },
  );
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
        select: (data) =>
          getDashboardNameFromReport(data.resource?.report?.spec),
      },
    },
  );
}

export function useReportOwnerName(
  organization: string,
  project: string,
  ownerId: string,
) {
  return createAdminServiceSearchProjectUsers(
    organization,
    project,
    {
      emailQuery: "%",
      pageSize: 1000,
      pageToken: undefined,
    },
    {
      query: {
        select: (data) => data.users.find((u) => u.id === ownerId)?.displayName,
      },
    },
  );
}

export function useIsReportCreatedByCode(instanceId: string, name: string) {
  return createRuntimeServiceGetResource(
    instanceId,
    {
      "name.name": name,
      "name.kind": ResourceKind.Report,
    },
    {
      query: {
        select: (data) =>
          !data.resource.report.spec.annotations["admin_owner_user_id"],
      },
    },
  );
}
