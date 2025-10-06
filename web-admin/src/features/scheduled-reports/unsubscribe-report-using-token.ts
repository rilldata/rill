/**
 * This file implements a variant of the `UnsubscribeReport` client code that authenticates using a bearer token.
 *
 * Modifications from the original Orval-generated code in `/web-admin/src/client/gen/admin-service/admin-service.ts` include:
 * - `queryFn`: Authentication via `Authorization: Bearer ${token}` header, replacing cookie-based authentication.
 * - `queryOptions`: Conditional enabling of the query based on the presence of `token`.
 */

import {
  createMutation,
  type CreateMutationOptions,
  type MutationFunction,
} from "@tanstack/svelte-query";
import {
  type AdminServiceUnsubscribeAlertBodyBody,
  type RpcStatus,
  type V1UnsubscribeReportResponse,
} from "@rilldata/web-admin/client";
import httpClient from "@rilldata/web-admin/client/http-client";

const adminServiceUnsubscribeReportWithToken = (
  organization: string,
  project: string,
  name: string,
  adminServiceUnsubscribeReportBody: AdminServiceUnsubscribeAlertBodyBody,
  token: string,
) => {
  return httpClient<V1UnsubscribeReportResponse>({
    url: `/v1/orgs/${organization}/projects/${project}/reports/${name}/unsubscribe`,
    method: "post",
    data: adminServiceUnsubscribeReportBody,
    // We use the bearer token to authenticate the request
    headers: {
      Authorization: `Bearer ${token}`,
    },
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
  });
};

export const createAdminServiceUnsubscribeReportUsingToken = <
  TError = RpcStatus,
  TContext = unknown,
>(options?: {
  mutation?: CreateMutationOptions<
    Awaited<ReturnType<typeof adminServiceUnsubscribeReportWithToken>>,
    TError,
    {
      organization: string;
      project: string;
      name: string;
      data: AdminServiceUnsubscribeAlertBodyBody;
      token: string;
    },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof adminServiceUnsubscribeReportWithToken>>,
    {
      organization: string;
      project: string;
      name: string;
      data: AdminServiceUnsubscribeAlertBodyBody;
      token: string;
    }
  > = (props) => {
    const { organization, project, name, token, data } = props ?? {};

    return adminServiceUnsubscribeReportWithToken(
      organization,
      project,
      name,
      data,
      token,
    );
  };

  return createMutation<
    Awaited<ReturnType<typeof adminServiceUnsubscribeReportWithToken>>,
    TError,
    {
      organization: string;
      project: string;
      name: string;
      data: AdminServiceUnsubscribeAlertBodyBody;
      token: string;
    },
    TContext
  >({ mutationFn, ...mutationOptions });
};
