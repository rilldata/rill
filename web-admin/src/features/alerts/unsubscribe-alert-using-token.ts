/**
 * This file implements a variant of the `UnsubscribeAlert` client code that authenticates using a bearer token.
 *
 * Modifications from the original Orval-generated code in `/web-admin/src/client/gen/default/default.ts` include:
 * - Request function: Authentication via `Authorization: Bearer ${token}` header, replacing cookie-based authentication.
 * - Mutation creator: It delegates to the generated creator, overriding only `mutationFn`, so the mutation key and variables match the generated code.
 */

import type { QueryClient } from "@tanstack/svelte-query";
import {
  createAdminServiceUnsubscribeAlert,
  type AdminServiceUnsubscribeAlertBodyBody,
  type V1UnsubscribeAlertResponse,
} from "@rilldata/web-admin/client";
import httpClient from "@rilldata/web-admin/client/http-client";

const adminServiceUnsubscribeAlertUsingToken = (
  org: string,
  project: string,
  name: string,
  adminServiceUnsubscribeAlertBodyBody: AdminServiceUnsubscribeAlertBodyBody,
  token: string,
) => {
  return httpClient<V1UnsubscribeAlertResponse>({
    url: `/v1/orgs/${org}/projects/${project}/alerts/${name}/unsubscribe`,
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    data: adminServiceUnsubscribeAlertBodyBody,
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
  });
};

export const createAdminServiceUnsubscribeAlertUsingToken = (
  token: string,
  queryClient?: QueryClient,
) =>
  createAdminServiceUnsubscribeAlert(
    {
      mutation: {
        mutationFn: ({ org, project, name, data }) =>
          adminServiceUnsubscribeAlertUsingToken(
            org,
            project,
            name,
            data,
            token,
          ),
      },
    },
    queryClient,
  );
