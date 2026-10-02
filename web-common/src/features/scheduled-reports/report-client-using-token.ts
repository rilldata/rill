/**
 * This file implements variants of the `CreateReport`, `EditReport` and `UnsubscribeReport` client code that authenticate using a bearer token.
 *
 * Modifications from the original Orval-generated code in `/web-admin/src/client/gen/default/default.ts` include:
 * - Request functions: Authentication via `Authorization: Bearer ${token}` header, replacing cookie-based authentication.
 * - Mutation creators: The token is bound when creating the mutation, so the mutation variables match the generated mutations.
 */

import {
  createMutation,
  type CreateMutationOptions,
  type CreateMutationResult,
  type QueryClient,
} from "@tanstack/svelte-query";
import {
  type AdminServiceCreateReportBodyBody,
  type AdminServiceUnsubscribeAlertBodyBody,
  type RpcStatus,
  type V1CreateReportResponse,
  type V1EditReportResponse,
  type V1UnsubscribeReportResponse,
} from "@rilldata/web-admin/client";
import httpClient from "@rilldata/web-admin/client/http-client.ts";

export const adminServiceCreateReportUsingToken = (
  org: string,
  project: string,
  adminServiceCreateReportBodyBody: AdminServiceCreateReportBodyBody,
  token: string,
) => {
  return httpClient<V1CreateReportResponse>({
    url: `/v1/orgs/${org}/projects/${project}/reports`,
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    data: adminServiceCreateReportBodyBody,
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
  });
};

export const createAdminServiceCreateReportUsingToken = <
  TError = RpcStatus,
  TContext = unknown,
>(
  token: string,
  options?: {
    mutation?: CreateMutationOptions<
      Awaited<ReturnType<typeof adminServiceCreateReportUsingToken>>,
      TError,
      { org: string; project: string; data: AdminServiceCreateReportBodyBody },
      TContext
    >;
  },
  queryClient?: QueryClient,
): CreateMutationResult<
  Awaited<ReturnType<typeof adminServiceCreateReportUsingToken>>,
  TError,
  { org: string; project: string; data: AdminServiceCreateReportBodyBody },
  TContext
> => {
  return createMutation(
    {
      mutationKey: ["adminServiceCreateReport"],
      mutationFn: ({ org, project, data }) =>
        adminServiceCreateReportUsingToken(org, project, data, token),
      ...options?.mutation,
    },
    queryClient,
  );
};

export const adminServiceEditReportUsingToken = (
  org: string,
  project: string,
  name: string,
  adminServiceCreateReportBodyBody: AdminServiceCreateReportBodyBody,
  token: string,
) => {
  return httpClient<V1EditReportResponse>({
    url: `/v1/orgs/${org}/projects/${project}/reports/${name}`,
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    data: adminServiceCreateReportBodyBody,
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
  });
};

export const createAdminServiceEditReportUsingToken = <
  TError = RpcStatus,
  TContext = unknown,
>(
  token: string,
  options?: {
    mutation?: CreateMutationOptions<
      Awaited<ReturnType<typeof adminServiceEditReportUsingToken>>,
      TError,
      {
        org: string;
        project: string;
        name: string;
        data: AdminServiceCreateReportBodyBody;
      },
      TContext
    >;
  },
  queryClient?: QueryClient,
): CreateMutationResult<
  Awaited<ReturnType<typeof adminServiceEditReportUsingToken>>,
  TError,
  {
    org: string;
    project: string;
    name: string;
    data: AdminServiceCreateReportBodyBody;
  },
  TContext
> => {
  return createMutation(
    {
      mutationKey: ["adminServiceEditReport"],
      mutationFn: ({ org, project, name, data }) =>
        adminServiceEditReportUsingToken(org, project, name, data, token),
      ...options?.mutation,
    },
    queryClient,
  );
};

export const adminServiceUnsubscribeReportUsingToken = (
  org: string,
  project: string,
  name: string,
  adminServiceUnsubscribeAlertBodyBody: AdminServiceUnsubscribeAlertBodyBody,
  token: string,
) => {
  return httpClient<V1UnsubscribeReportResponse>({
    url: `/v1/orgs/${org}/projects/${project}/reports/${name}/unsubscribe`,
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

export const createAdminServiceUnsubscribeReportUsingToken = <
  TError = RpcStatus,
  TContext = unknown,
>(
  token: string,
  options?: {
    mutation?: CreateMutationOptions<
      Awaited<ReturnType<typeof adminServiceUnsubscribeReportUsingToken>>,
      TError,
      {
        org: string;
        project: string;
        name: string;
        data: AdminServiceUnsubscribeAlertBodyBody;
      },
      TContext
    >;
  },
  queryClient?: QueryClient,
): CreateMutationResult<
  Awaited<ReturnType<typeof adminServiceUnsubscribeReportUsingToken>>,
  TError,
  {
    org: string;
    project: string;
    name: string;
    data: AdminServiceUnsubscribeAlertBodyBody;
  },
  TContext
> => {
  return createMutation(
    {
      mutationKey: ["adminServiceUnsubscribeReport"],
      mutationFn: ({ org, project, name, data }) =>
        adminServiceUnsubscribeReportUsingToken(
          org,
          project,
          name,
          data,
          token,
        ),
      ...options?.mutation,
    },
    queryClient,
  );
};
