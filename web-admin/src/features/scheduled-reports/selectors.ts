import { createAdminServiceSearchProjectUsers } from "@rilldata/web-admin/client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { getDashboardNameFromReport } from "@rilldata/web-common/features/scheduled-reports/utils";
import type { V1ReportSpec } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import {
  createRuntimeServiceGetResource,
  createRuntimeServiceListResources,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { smartRefetchIntervalFunc } from "@rilldata/web-admin/lib/refetch-interval-store";

export function useReports(client: RuntimeClient, enabled = true) {
  return createRuntimeServiceListResources(
    client,
    {
      kind: ResourceKind.Report,
    },
    {
      query: {
        enabled: enabled && !!client.instanceId,
        refetchOnMount: true,
        refetchInterval: smartRefetchIntervalFunc,
      },
    },
  );
}

export function useReport(client: RuntimeClient, name: string) {
  return createRuntimeServiceGetResource(client, {
    name: { name, kind: ResourceKind.Report },
  });
}

export function useReportDashboardName(client: RuntimeClient, name: string) {
  return createRuntimeServiceGetResource(
    client,
    {
      name: { name, kind: ResourceKind.Report },
    },
    {
      query: {
        select: (data) =>
          getDashboardNameFromReport(
            data.resource?.report?.spec as V1ReportSpec,
          ),
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

export function useIsReportCreatedByCode(client: RuntimeClient, name: string) {
  return createRuntimeServiceGetResource(
    client,
    {
      name: { name, kind: ResourceKind.Report },
    },
    {
      query: {
        select: (data) =>
          !data.resource.report.spec.annotations["admin_owner_user_id"],
      },
    },
  );
}
