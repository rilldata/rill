import { createAdminServiceSearchProjectUsers } from "@rilldata/web-admin/client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { getDashboardNameFromReport } from "@rilldata/web-common/features/scheduled-reports/utils";
import {
  createRuntimeServiceGetResource,
  createRuntimeServiceListResources,
  type V1ListResourcesResponse,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { Query } from "@tanstack/svelte-query";

const INITIAL_REFETCH_INTERVAL = 200; // Start at 200ms for immediate feedback
const MAX_REFETCH_INTERVAL = 2_000; // Cap at 2s
const BACKOFF_FACTOR = 1.5;

function isResourceReconciling(resource: V1Resource) {
  return (
    resource.meta.reconcileStatus === "RECONCILE_STATUS_PENDING" ||
    resource.meta.reconcileStatus === "RECONCILE_STATUS_RUNNING"
  );
}

function calculateRefetchInterval(
  currentInterval: number,
  data: V1ListResourcesResponse | undefined,
  query: Query,
): number | false {
  if (query.state.error) return false;
  if (!data?.resources) return INITIAL_REFETCH_INTERVAL;

  const hasReconcilingResources = data.resources.some(isResourceReconciling);

  if (!hasReconcilingResources) {
    return false;
  }

  return Math.min(currentInterval * BACKOFF_FACTOR, MAX_REFETCH_INTERVAL);
}

export function useReports(instanceId: string, enabled = true) {
  let currentRefetchInterval = INITIAL_REFETCH_INTERVAL;

  return createRuntimeServiceListResources(
    instanceId,
    {
      kind: ResourceKind.Report,
    },
    {
      query: {
        enabled: enabled && !!instanceId,
        refetchOnMount: true,
        refetchInterval: (query) => {
          const newInterval = calculateRefetchInterval(
            currentRefetchInterval,
            query.state.data,
            query,
          );
          if (newInterval === false) {
            currentRefetchInterval = INITIAL_REFETCH_INTERVAL;
            return false;
          }
          currentRefetchInterval = newInterval;
          return newInterval;
        },
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
