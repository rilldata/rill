import { createAdminServiceSearchProjectUsers } from "@rilldata/web-admin/client";
import { getExploreName } from "@rilldata/web-admin/features/dashboards/query-mappers/utils";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  createRuntimeServiceGetResource,
  createRuntimeServiceListResources,
} from "@rilldata/web-common/runtime-client";

export function useAlerts(instanceId: string, enabled = true) {
  return createRuntimeServiceListResources(
    instanceId,
    {
      kind: ResourceKind.Alert,
    },
    {
      query: {
        enabled: enabled && !!instanceId,
        refetchOnMount: true,
      },
    },
  );
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
          const alertSpec = data.resource?.alert?.spec;
          if (!alertSpec) return "";

          if (alertSpec.annotations.web_open_path)
            return getExploreName(alertSpec.annotations.web_open_path);

          const queryArgsJson = JSON.parse(
            alertSpec.resolverProperties.query_args_json ||
              alertSpec.queryArgsJson ||
              "{}",
          );

          return (
            queryArgsJson?.metrics_view_name ??
            queryArgsJson?.metricsViewName ??
            queryArgsJson?.metricsView
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
