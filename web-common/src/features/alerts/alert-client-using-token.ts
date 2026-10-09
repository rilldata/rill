/**
 * This file implements variants of the `CreateAlert`, `EditAlert` and `UnsubscribeAlert` client code that authenticate using a bearer token.
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
  type AdminServiceCreateAlertBodyBody,
  type AdminServiceUnsubscribeAlertBodyBody,
  type RpcStatus,
  type V1CreateAlertResponse,
  type V1EditAlertResponse,
  type V1UnsubscribeAlertResponse,
} from "@rilldata/web-admin/client";
import httpClient from "@rilldata/web-admin/client/http-client";

export const adminServiceCreateAlertUsingToken = (
  org: string,
  project: string,
  adminServiceCreateAlertBodyBody: AdminServiceCreateAlertBodyBody,
  token: string,
) => {
  return httpClient<V1CreateAlertResponse>({
    url: `/v1/orgs/${org}/projects/${project}/alerts`,
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    data: adminServiceCreateAlertBodyBody,
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
  });
};

export const createAdminServiceCreateAlertUsingToken = <
  TError = RpcStatus,
  TContext = unknown,
>(
  token: string,
  options?: {
    mutation?: CreateMutationOptions<
      Awaited<ReturnType<typeof adminServiceCreateAlertUsingToken>>,
      TError,
      { org: string; project: string; data: AdminServiceCreateAlertBodyBody },
      TContext
    >;
  },
  queryClient?: QueryClient,
): CreateMutationResult<
  Awaited<ReturnType<typeof adminServiceCreateAlertUsingToken>>,
  TError,
  { org: string; project: string; data: AdminServiceCreateAlertBodyBody },
  TContext
> => {
  return createMutation(
    {
      mutationKey: ["adminServiceCreateAlert"],
      mutationFn: ({ org, project, data }) =>
        adminServiceCreateAlertUsingToken(org, project, data, token),
      ...options?.mutation,
    },
    queryClient,
  );
};

export const adminServiceEditAlertUsingToken = (
  org: string,
  project: string,
  name: string,
  adminServiceCreateAlertBodyBody: AdminServiceCreateAlertBodyBody,
  token: string,
) => {
  return httpClient<V1EditAlertResponse>({
    url: `/v1/orgs/${org}/projects/${project}/alerts/${name}`,
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    data: adminServiceCreateAlertBodyBody,
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
  });
};

export const createAdminServiceEditAlertUsingToken = <
  TError = RpcStatus,
  TContext = unknown,
>(
  token: string,
  options?: {
    mutation?: CreateMutationOptions<
      Awaited<ReturnType<typeof adminServiceEditAlertUsingToken>>,
      TError,
      {
        org: string;
        project: string;
        name: string;
        data: AdminServiceCreateAlertBodyBody;
      },
      TContext
    >;
  },
  queryClient?: QueryClient,
): CreateMutationResult<
  Awaited<ReturnType<typeof adminServiceEditAlertUsingToken>>,
  TError,
  {
    org: string;
    project: string;
    name: string;
    data: AdminServiceCreateAlertBodyBody;
  },
  TContext
> => {
  return createMutation(
    {
      mutationKey: ["adminServiceEditAlert"],
      mutationFn: ({ org, project, name, data }) =>
        adminServiceEditAlertUsingToken(org, project, name, data, token),
      ...options?.mutation,
    },
    queryClient,
  );
};

export const adminServiceUnsubscribeAlertUsingToken = (
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

export const createAdminServiceUnsubscribeAlertUsingToken = <
  TError = RpcStatus,
  TContext = unknown,
>(
  token: string,
  options?: {
    mutation?: CreateMutationOptions<
      Awaited<ReturnType<typeof adminServiceUnsubscribeAlertUsingToken>>,
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
  Awaited<ReturnType<typeof adminServiceUnsubscribeAlertUsingToken>>,
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
      mutationKey: ["adminServiceUnsubscribeAlert"],
      mutationFn: ({ org, project, name, data }) =>
        adminServiceUnsubscribeAlertUsingToken(org, project, name, data, token),
      ...options?.mutation,
    },
    queryClient,
  );
};
