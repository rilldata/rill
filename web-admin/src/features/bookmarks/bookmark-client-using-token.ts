/**
 * This file implements variants of the bookmark client code that authenticate using a bearer token.
 * Embeds use them with the embed's admin token, since they have no cookie-based session with the admin server.
 *
 * Modifications from the original Orval-generated code in `/web-admin/src/client/gen/default/default.ts` include:
 * - Request functions: Authentication via `Authorization: Bearer ${token}` header, replacing cookie-based authentication.
 * - Query options: The token is bound when creating them; the query key matches the generated code, so cache invalidation is shared between the two variants.
 * - Mutation creators: They delegate to the generated creators, overriding only `mutationFn`, so mutation keys and variables match the generated code.
 */

import type { CreateQueryOptions, QueryClient } from "@tanstack/svelte-query";
import {
  createAdminServiceCreateBookmark,
  createAdminServiceRemoveBookmark,
  createAdminServiceUpdateBookmark,
  getAdminServiceListBookmarksQueryKey,
  type AdminServiceListBookmarksParams,
  type RpcStatus,
  type V1CreateBookmarkRequest,
  type V1CreateBookmarkResponse,
  type V1ListBookmarksResponse,
  type V1RemoveBookmarkResponse,
  type V1UpdateBookmarkRequest,
  type V1UpdateBookmarkResponse,
} from "@rilldata/web-admin/client";
import httpClient from "@rilldata/web-admin/client/http-client.ts";

export const adminServiceListBookmarksUsingToken = (
  params: AdminServiceListBookmarksParams,
  token: string,
  signal?: AbortSignal,
) => {
  return httpClient<V1ListBookmarksResponse>({
    url: `/v1/users/bookmarks`,
    method: "GET",
    params,
    headers: {
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
    signal,
  });
};

export const getAdminServiceListBookmarksUsingTokenQueryOptions = (
  params: AdminServiceListBookmarksParams,
  token: string,
): CreateQueryOptions<V1ListBookmarksResponse, RpcStatus> => {
  return {
    queryKey: getAdminServiceListBookmarksQueryKey(params),
    queryFn: ({ signal }) =>
      adminServiceListBookmarksUsingToken(params, token, signal),
    enabled: !!token,
  };
};

export const adminServiceCreateBookmarkUsingToken = (
  v1CreateBookmarkRequest: V1CreateBookmarkRequest,
  token: string,
) => {
  return httpClient<V1CreateBookmarkResponse>({
    url: `/v1/users/bookmarks`,
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    data: v1CreateBookmarkRequest,
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
  });
};

export const createAdminServiceCreateBookmarkUsingToken = (
  token: string,
  queryClient?: QueryClient,
) =>
  createAdminServiceCreateBookmark(
    {
      mutation: {
        mutationFn: ({ data }) =>
          adminServiceCreateBookmarkUsingToken(data, token),
      },
    },
    queryClient,
  );

export const adminServiceUpdateBookmarkUsingToken = (
  v1UpdateBookmarkRequest: V1UpdateBookmarkRequest,
  token: string,
) => {
  return httpClient<V1UpdateBookmarkResponse>({
    url: `/v1/users/bookmarks`,
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    data: v1UpdateBookmarkRequest,
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
  });
};

export const createAdminServiceUpdateBookmarkUsingToken = (
  token: string,
  queryClient?: QueryClient,
) =>
  createAdminServiceUpdateBookmark(
    {
      mutation: {
        mutationFn: ({ data }) =>
          adminServiceUpdateBookmarkUsingToken(data, token),
      },
    },
    queryClient,
  );

export const adminServiceRemoveBookmarkUsingToken = (
  bookmarkId: string,
  token: string,
) => {
  return httpClient<V1RemoveBookmarkResponse>({
    url: `/v1/users/bookmarks/${bookmarkId}`,
    method: "DELETE",
    headers: {
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
  });
};

export const createAdminServiceRemoveBookmarkUsingToken = (
  token: string,
  queryClient?: QueryClient,
) =>
  createAdminServiceRemoveBookmark(
    {
      mutation: {
        mutationFn: ({ bookmarkId }) =>
          adminServiceRemoveBookmarkUsingToken(bookmarkId, token),
      },
    },
    queryClient,
  );
