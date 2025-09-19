import type { V1GetOrganizationResponse } from "@rilldata/web-admin/client";
import httpClient from "@rilldata/web-admin/client/http-client";

export const getOrgWithBearerToken = (
  organizationName: string,
  token: string,
) => {
  return httpClient<V1GetOrganizationResponse>({
    url: `/v1/orgs/${organizationName}`,
    method: "get",
    // We use the bearer token to authenticate the request
    headers: {
      Authorization: `Bearer ${token}`,
    },
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
  });
};
