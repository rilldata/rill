import {
  adminServiceGetOrganization,
  getAdminServiceGetOrganizationQueryKey,
  type V1GetOrganizationResponse,
  type V1Organization,
} from "@rilldata/web-admin/client";
import { listProjectsForOrgQueryOptions } from "@rilldata/web-admin/features/projects/list-projects-query-options";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { FetchQueryOptions } from "@tanstack/query-core";
import { createQuery } from "@tanstack/svelte-query";

export function areAllProjectsHibernating(organization: string) {
  return createQuery({
    ...listProjectsForOrgQueryOptions(organization),
    select: (data) =>
      data.projects?.length &&
      data.projects.every((p) => !p.primaryDeploymentId),
  });
}

export async function fetchAllProjectsHibernating(organization: string) {
  const projectsResp = await queryClient.fetchQuery(
    listProjectsForOrgQueryOptions(organization),
  );
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
