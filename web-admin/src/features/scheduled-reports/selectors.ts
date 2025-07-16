import { createAdminServiceSearchProjectUsers } from "@rilldata/web-admin/client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { getDashboardNameFromReport } from "@rilldata/web-common/features/scheduled-reports/utils";
import {
  createRuntimeServiceGetResource,
  createRuntimeServiceListResources,
} from "@rilldata/web-common/runtime-client";
import {
  updateSmartRefetchMeta,
  INITIAL_REFETCH_INTERVAL,
} from "../shared/refetch-interval-store";
import { writable, get } from "svelte/store";

export function useReports(instanceId: string, enabled = true) {
  // Local store for per-query refetch interval
  const refetchIntervalStore = writable<number | false>(
    INITIAL_REFETCH_INTERVAL,
  );

  return createRuntimeServiceListResources(
    instanceId,
    {
      kind: ResourceKind.Report,
    },
    {
      query: {
        enabled: enabled && !!instanceId,
        refetchOnMount: true,
        select: (data) => {
          // Update the local refetch interval store
          const meta = updateSmartRefetchMeta(data?.resources, {
            refetchInterval: get(refetchIntervalStore),
          });
          refetchIntervalStore.set(meta.refetchInterval);
          return data;
        },
        refetchInterval: () => get(refetchIntervalStore),
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
