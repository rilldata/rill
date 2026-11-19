import {
  createAdminServiceGetEmbedUser,
  createAdminServiceSearchProjectUsers,
} from "@rilldata/web-admin/client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { getDashboardNameFromReport } from "@rilldata/web-common/features/scheduled-reports/utils";
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
          getDashboardNameFromReport(data.resource?.report?.spec ?? {}),
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
        select: (data) =>
          data.users?.find((u) => u.id === ownerId)?.displayName,
        enabled: !!ownerId,
      },
    },
  );
}

// useEmbedUserExternalId returns the external user ID of the embed user who created a report from an embed.
// It is only available to users who can manage reports.
export function useEmbedUserExternalId(
  organization: string,
  project: string,
  embedUserId: string,
) {
  return createAdminServiceGetEmbedUser(organization, project, embedUserId, {
    query: {
      select: (data) => data.externalUserId,
      enabled: !!embedUserId,
    },
  });
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
          !data.resource?.report?.spec?.annotations?.["admin_owner_user_id"] &&
          !data.resource?.report?.spec?.annotations?.["admin_embed_owner_id"],
      },
    },
  );
}
