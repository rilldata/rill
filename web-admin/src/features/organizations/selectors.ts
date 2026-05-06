import {
  adminServiceGetOrganization,
  getAdminServiceGetOrganizationQueryKey,
  type V1GetOrganizationResponse,
  type V1Organization,
} from "@rilldata/web-admin/client";
import type { FetchQueryOptions } from "@tanstack/query-core";

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
