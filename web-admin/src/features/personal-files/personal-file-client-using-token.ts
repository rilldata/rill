/**
 * This file implements variants of the personal file client code that authenticate using a bearer token.
 * Embeds use them with the embed's admin token, since they have no cookie-based session with the admin server.
 *
 * Modifications from the original Orval-generated code in `/web-admin/src/client/gen/default/default.ts` include:
 * - Request functions: Authentication via `Authorization: Bearer ${token}` header, replacing cookie-based authentication.
 * - Query creators: The token is bound when creating them; query keys match the generated code, so cache invalidation is shared between the two variants.
 * - Mutation creators: They delegate to the generated creators, overriding only `mutationFn`, so mutation keys and variables match the generated code.
 */

import { createQuery, type QueryClient } from "@tanstack/svelte-query";
import {
  createAdminServiceCreatePersonalFile,
  createAdminServiceDeletePersonalFile,
  getAdminServiceListPersonalFilesQueryKey,
  type AdminServiceCreatePersonalFileBody,
  type AdminServiceEditPersonalFileBody,
  type V1CreatePersonalFileResponse,
  type V1DeletePersonalFileResponse,
  type V1EditPersonalFileResponse,
  type V1GetPersonalFileResponse,
  type V1ListPersonalFilesResponse,
} from "@rilldata/web-admin/client";
import httpClient from "@rilldata/web-admin/client/http-client.ts";

export const adminServiceListPersonalFilesUsingToken = (
  org: string,
  project: string,
  token: string,
  signal?: AbortSignal,
) => {
  return httpClient<V1ListPersonalFilesResponse>({
    url: `/v1/orgs/${org}/projects/${project}/personal-files`,
    method: "GET",
    headers: {
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
    signal,
  });
};

export const createAdminServiceListPersonalFilesUsingToken = (
  org: string,
  project: string,
  token: string,
  queryClient?: QueryClient,
) => {
  return createQuery(
    {
      queryKey: getAdminServiceListPersonalFilesQueryKey(org, project),
      queryFn: ({ signal }) =>
        adminServiceListPersonalFilesUsingToken(org, project, token, signal),
      enabled: !!(org && project),
    },
    queryClient,
  );
};

export const adminServiceCreatePersonalFileUsingToken = (
  org: string,
  project: string,
  adminServiceCreatePersonalFileBody: AdminServiceCreatePersonalFileBody,
  token: string,
) => {
  return httpClient<V1CreatePersonalFileResponse>({
    url: `/v1/orgs/${org}/projects/${project}/personal-files`,
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    data: adminServiceCreatePersonalFileBody,
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
  });
};

export const createAdminServiceCreatePersonalFileUsingToken = (
  token: string,
  queryClient?: QueryClient,
) =>
  createAdminServiceCreatePersonalFile(
    {
      mutation: {
        mutationFn: ({ org, project, data }) =>
          adminServiceCreatePersonalFileUsingToken(org, project, data, token),
      },
    },
    queryClient,
  );

export const adminServiceGetPersonalFileUsingToken = (
  org: string,
  project: string,
  name: string,
  token: string,
  signal?: AbortSignal,
) => {
  return httpClient<V1GetPersonalFileResponse>({
    url: `/v1/orgs/${org}/projects/${project}/personal-files/${name}`,
    method: "GET",
    headers: {
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
    signal,
  });
};

export const adminServiceEditPersonalFileUsingToken = (
  org: string,
  project: string,
  name: string,
  adminServiceEditPersonalFileBody: AdminServiceEditPersonalFileBody,
  token: string,
) => {
  return httpClient<V1EditPersonalFileResponse>({
    url: `/v1/orgs/${org}/projects/${project}/personal-files/${name}`,
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    data: adminServiceEditPersonalFileBody,
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
  });
};

export const adminServiceDeletePersonalFileUsingToken = (
  org: string,
  project: string,
  name: string,
  token: string,
) => {
  return httpClient<V1DeletePersonalFileResponse>({
    url: `/v1/orgs/${org}/projects/${project}/personal-files/${name}`,
    method: "DELETE",
    headers: {
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
  });
};

export const createAdminServiceDeletePersonalFileUsingToken = (
  token: string,
  queryClient?: QueryClient,
) =>
  createAdminServiceDeletePersonalFile(
    {
      mutation: {
        mutationFn: ({ org, project, name }) =>
          adminServiceDeletePersonalFileUsingToken(org, project, name, token),
      },
    },
    queryClient,
  );
