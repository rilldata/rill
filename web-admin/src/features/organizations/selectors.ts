import {
  adminServiceGetOrganization,
  adminServiceListProjectsForOrganization,
  createAdminServiceListProjectsForOrganization,
  getAdminServiceGetOrganizationQueryKey,
  getAdminServiceListProjectsForOrganizationQueryKey,
  type V1GetOrganizationResponse,
  type V1Organization,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { FetchQueryOptions } from "@tanstack/query-core";

export function areAllProjectsHibernating(organization: string) {
  return createAdminServiceListProjectsForOrganization(
    organization,
    undefined,
    {
      query: {
        enabled: !!organization,
        select: (data) =>
          data.projects?.length &&
          data.projects.every((p) => !p.primaryDeploymentId),
      },
    },
  );
}

export async function fetchAllProjectsHibernating(organization: string) {
  const projectsResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceListProjectsForOrganizationQueryKey(organization),
    queryFn: () => adminServiceListProjectsForOrganization(organization),
    staleTime: Infinity,
  });
  return projectsResp.projects?.every((p) => !p.primaryDeploymentId) ?? false;
}

function normalizeOrganization(
  organization: string | V1Organization | undefined,
): string {
  if (typeof organization === "string") {
    return organization;
  }
  if (
    organization &&
    typeof organization === "object" &&
    "name" in organization &&
    typeof organization.name === "string"
  ) {
    return organization.name;
  }
  throw new Error(
    `Invalid organization parameter: expected string or V1Organization object with name property, got ${typeof organization}`,
  );
}

export function getFetchOrganizationQueryOptions(
  organization: string | V1Organization | undefined,
) {
  const orgName = normalizeOrganization(organization);
  return <FetchQueryOptions<V1GetOrganizationResponse>>{
    queryKey: getAdminServiceGetOrganizationQueryKey(orgName),
    queryFn: () => adminServiceGetOrganization(orgName),
    staleTime: Infinity,
  };
}
