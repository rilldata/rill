/**
 * This file implements variants of the bookmark client code that authenticate using a bearer token.
 * Embeds use them with the embed's admin token, since they have no cookie-based session with the admin server.
 *
 * Modifications from the original Orval-generated code in `/web-admin/src/client/gen/default/default.ts` include:
 * - Request functions: Authentication via `Authorization: Bearer ${token}` header, replacing cookie-based authentication.
 * - Query and mutation creators: The token is bound when creating them; query keys and mutation variables match the generated code,
 *   so cache invalidation is shared between the two variants.
 */

import {
  createMutation,
  type CreateMutationOptions,
  type CreateMutationResult,
  type CreateQueryOptions,
  type QueryClient,
} from "@tanstack/svelte-query";
import {
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

export const createAdminServiceCreateBookmarkUsingToken = <
  TError = RpcStatus,
  TContext = unknown,
>(
  token: string,
  options?: {
    mutation?: CreateMutationOptions<
      Awaited<ReturnType<typeof adminServiceCreateBookmarkUsingToken>>,
      TError,
      { data: V1CreateBookmarkRequest },
      TContext
    >;
  },
  queryClient?: QueryClient,
): CreateMutationResult<
  Awaited<ReturnType<typeof adminServiceCreateBookmarkUsingToken>>,
  TError,
  { data: V1CreateBookmarkRequest },
  TContext
> => {
  return createMutation(
    {
      mutationKey: ["adminServiceCreateBookmark"],
      mutationFn: ({ data }) =>
        adminServiceCreateBookmarkUsingToken(data, token),
      ...options?.mutation,
    },
    queryClient,
  );
};

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

export const createAdminServiceUpdateBookmarkUsingToken = <
  TError = RpcStatus,
  TContext = unknown,
>(
  token: string,
  options?: {
    mutation?: CreateMutationOptions<
      Awaited<ReturnType<typeof adminServiceUpdateBookmarkUsingToken>>,
      TError,
      { data: V1UpdateBookmarkRequest },
      TContext
    >;
  },
  queryClient?: QueryClient,
): CreateMutationResult<
  Awaited<ReturnType<typeof adminServiceUpdateBookmarkUsingToken>>,
  TError,
  { data: V1UpdateBookmarkRequest },
  TContext
> => {
  return createMutation(
    {
      mutationKey: ["adminServiceUpdateBookmark"],
      mutationFn: ({ data }) =>
        adminServiceUpdateBookmarkUsingToken(data, token),
      ...options?.mutation,
    },
    queryClient,
  );
};

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

export const createAdminServiceRemoveBookmarkUsingToken = <
  TError = RpcStatus,
  TContext = unknown,
>(
  token: string,
  options?: {
    mutation?: CreateMutationOptions<
      Awaited<ReturnType<typeof adminServiceRemoveBookmarkUsingToken>>,
      TError,
      { bookmarkId: string },
      TContext
    >;
  },
  queryClient?: QueryClient,
): CreateMutationResult<
  Awaited<ReturnType<typeof adminServiceRemoveBookmarkUsingToken>>,
  TError,
  { bookmarkId: string },
  TContext
> => {
  return createMutation(
    {
      mutationKey: ["adminServiceRemoveBookmark"],
      mutationFn: ({ bookmarkId }) =>
        adminServiceRemoveBookmarkUsingToken(bookmarkId, token),
      ...options?.mutation,
    },
    queryClient,
  );
};
