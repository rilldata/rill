import { createAdminServiceSearchProjectUsers } from "@rilldata/web-admin/client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  createRuntimeServiceGetResource,
  createRuntimeServiceListResources,
} from "@rilldata/web-common/runtime-client";

export function useAlerts(instanceId: string) {
  return createRuntimeServiceListResources(instanceId, {
    kind: ResourceKind.Alert,
  });
}

export function useAlert(instanceId: string, name: string) {
  return createRuntimeServiceGetResource(instanceId, {
    "name.name": name,
    "name.kind": ResourceKind.Alert,
  });
}

export function useAlertDashboardName(instanceId: string, name: string) {
  return createRuntimeServiceGetResource(
    instanceId,
    {
      "name.name": name,
      "name.kind": ResourceKind.Alert,
    },
    {
      query: {
        select: (data) => {
          const queryArgsJson = JSON.parse(
            data.resource.alert.spec.queryArgsJson,
          );

          return (
            queryArgsJson?.metrics_view_name ??
            queryArgsJson?.metricsViewName ??
            null
          );
        },
      },
    },
  );
}

export function useAlertOwnerName(
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

export function useIsAlertCreatedByCode(instanceId: string, name: string) {
  return createRuntimeServiceGetResource(
    instanceId,
    {
      "name.name": name,
      "name.kind": ResourceKind.Alert,
    },
    {
      query: {
        select: (data) =>
          !data.resource.alert.spec.annotations["admin_owner_user_id"],
      },
    },
  );
}
