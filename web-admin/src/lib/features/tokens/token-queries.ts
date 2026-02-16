import {
  createMutation,
  createQuery,
  useQueryClient,
} from "@tanstack/svelte-query";
import { adminServiceClient } from "../../client";
import type { V1ServiceToken } from "@rilldata/web-common/runtime-client";

/**
 * Query key factory for token-related queries.
 * Centralizes key construction for consistent cache management.
 */
export const tokenKeys = {
  serviceTokens: (orgName: string) => ["service-tokens", orgName] as const,
  serviceTokenList: (
    orgName: string,
    options?: { search?: string; pageToken?: string },
  ) =>
    [
      ...tokenKeys.serviceTokens(orgName),
      "list",
      options?.search ?? "",
      options?.pageToken ?? "",
    ] as const,
  serviceTokenDetail: (orgName: string, tokenId: string) =>
    [...tokenKeys.serviceTokens(orgName), "detail", tokenId] as const,

  userTokens: () => ["user-tokens"] as const,
  userTokenList: (options?: { pageToken?: string }) =>
    [...tokenKeys.userTokens(), "list", options?.pageToken ?? ""] as const,
  userTokenDetail: (tokenId: string) =>
    [...tokenKeys.userTokens(), "detail", tokenId] as const,
};

// ─── Service Token Queries ───────────────────────────────────────────────────

/**
 * Fetches a paginated list of service tokens for an organization.
 * Supports optional search filtering and cursor-based pagination.
 */
export function createServiceTokenListQuery(
  orgName: string,
  options?: { search?: string; pageToken?: string },
) {
  return createQuery({
    queryKey: tokenKeys.serviceTokenList(orgName, options),
    queryFn: async () => {
      const params: Record<string, string> = {};
      if (options?.search) {
        params.search = options.search;
      }
      if (options?.pageToken) {
        params.pageToken = options.pageToken;
      }
      params.pageSize = "20";

      const response =
        await adminServiceClient.adminServiceListServiceTokens(orgName, params);
      return response;
    },
    enabled: !!orgName,
    keepPreviousData: true,
  });
}

/**
 * Fetches detailed information about a single service token.
 */
export function createGetServiceTokenQuery(
  orgName: string,
  tokenId: string,
  options?: { enabled?: boolean },
) {
  return createQuery({
    queryKey: tokenKeys.serviceTokenDetail(orgName, tokenId),
    queryFn: async () => {
      const response = await adminServiceClient.adminServiceGetServiceToken(
        orgName,
        tokenId,
      );
      return response;
    },
    enabled: (options?.enabled ?? true) && !!orgName && !!tokenId,
  });
}

/**
 * Creates a new service token for an organization.
 * Automatically invalidates the service token list cache on success.
 *
 * The mutation returns the full response including the plaintext token,
 * which is only available at creation time.
 */
export function createServiceTokenMutation(orgName: string) {
  const queryClient = useQueryClient();

  return createMutation({
    mutationFn: async (input: {
      name: string;
      description?: string;
      scope?: "organization" | "project";
      projectId?: string;
      permissions?: string;
    }) => {
      const body: Record<string, unknown> = {
        name: input.name,
      };
      if (input.description) {
        body.description = input.description;
      }
      if (input.scope === "project" && input.projectId) {
        body.projectId = input.projectId;
      }
      if (input.permissions) {
        body.permissions = input.permissions;
      }

      const response = await adminServiceClient.adminServiceCreateServiceToken(
        orgName,
        body,
      );
      return response;
    },
    onSuccess: () => {
      // Invalidate all service token list queries for this org
      void queryClient.invalidateQueries(tokenKeys.serviceTokens(orgName));
    },
  });
}

/**
 * Deletes (revokes) a service token.
 * Automatically invalidates the service token list cache on success.
 */
export function createDeleteServiceTokenMutation(orgName: string) {
  const queryClient = useQueryClient();

  return createMutation({
    mutationFn: async (tokenId: string) => {
      await adminServiceClient.adminServiceDeleteServiceToken(
        orgName,
        tokenId,
      );
    },
    onSuccess: () => {
      // Invalidate all service token queries for this org (list + details)
      void queryClient.invalidateQueries(tokenKeys.serviceTokens(orgName));
    },
  });
}

// ─── User Token Queries ──────────────────────────────────────────────────────

/**
 * Fetches a paginated list of the current user's personal tokens.
 */
export function createUserTokenListQuery(options?: { pageToken?: string }) {
  return createQuery({
    queryKey: tokenKeys.userTokenList(options),
    queryFn: async () => {
      const params: Record<string, string> = {};
      if (options?.pageToken) {
        params.pageToken = options.pageToken;
      }
      params.pageSize = "20";

      const response =
        await adminServiceClient.adminServiceListUserTokens(params);
      return response;
    },
    keepPreviousData: true,
  });
}

/**
 * Fetches detailed information about a single user token.
 */
export function createGetUserTokenQuery(
  tokenId: string,
  options?: { enabled?: boolean },
) {
  return createQuery({
    queryKey: tokenKeys.userTokenDetail(tokenId),
    queryFn: async () => {
      const response =
        await adminServiceClient.adminServiceGetUserToken(tokenId);
      return response;
    },
    enabled: (options?.enabled ?? true) && !!tokenId,
  });
}

/**
 * Creates a new personal user token.
 * Automatically invalidates the user token list cache on success.
 *
 * The mutation returns the full response including the plaintext token,
 * which is only available at creation time.
 */
export function createUserTokenMutation() {
  const queryClient = useQueryClient();

  return createMutation({
    mutationFn: async (input: {
      name: string;
      description?: string;
      expiresInDays?: number;
      expiresAt?: string;
    }) => {
      const body: Record<string, unknown> = {
        name: input.name,
      };
      if (input.description) {
        body.description = input.description;
      }
      if (input.expiresAt) {
        body.expiresAt = input.expiresAt;
      } else if (input.expiresInDays) {
        // Compute expiration date from days
        const expiresAt = new Date();
        expiresAt.setDate(expiresAt.getDate() + input.expiresInDays);
        body.expiresAt = expiresAt.toISOString();
      }

      const response =
        await adminServiceClient.adminServiceCreateUserToken(body);
      return response;
    },
    onSuccess: () => {
      // Invalidate all user token queries (list + details)
      void queryClient.invalidateQueries(tokenKeys.userTokens());
    },
  });
}

/**
 * Deletes (revokes) a personal user token.
 * Automatically invalidates the user token list cache on success.
 */
export function createDeleteUserTokenMutation() {
  const queryClient = useQueryClient();

  return createMutation({
    mutationFn: async (tokenId: string) => {
      await adminServiceClient.adminServiceDeleteUserToken(tokenId);
    },
    onSuccess: () => {
      // Invalidate all user token queries (list + details)
      void queryClient.invalidateQueries(tokenKeys.userTokens());
    },
  });
}

// ─── Bulk Operations ─────────────────────────────────────────────────────────

/**
 * Deletes multiple service tokens in parallel.
 * Issues individual delete calls concurrently and invalidates the cache once all complete.
 * Returns a summary of successes and failures.
 */
export function createBulkDeleteServiceTokensMutation(orgName: string) {
  const queryClient = useQueryClient();

  return createMutation({
    mutationFn: async (tokenIds: string[]) => {
      const results = await Promise.allSettled(
        tokenIds.map((id) =>
          adminServiceClient.adminServiceDeleteServiceToken(orgName, id),
        ),
      );

      const succeeded = results.filter(
        (r) => r.status === "fulfilled",
      ).length;
      const failed = results.filter((r) => r.status === "rejected").length;

      if (failed > 0 && succeeded === 0) {
        throw new Error(`Failed to revoke all ${failed} tokens`);
      }

      return { succeeded, failed, total: tokenIds.length };
    },
    onSuccess: () => {
      void queryClient.invalidateQueries(tokenKeys.serviceTokens(orgName));
    },
    onError: () => {
      // Even on partial failure, invalidate to refresh the list
      void queryClient.invalidateQueries(tokenKeys.serviceTokens(orgName));
    },
  });
}

/**
 * Deletes multiple user tokens in parallel.
 * Issues individual delete calls concurrently and invalidates the cache once all complete.
 */
export function createBulkDeleteUserTokensMutation() {
  const queryClient = useQueryClient();

  return createMutation({
    mutationFn: async (tokenIds: string[]) => {
      const results = await Promise.allSettled(
        tokenIds.map((id) =>
          adminServiceClient.adminServiceDeleteUserToken(id),
        ),
      );

      const succeeded = results.filter(
        (r) => r.status === "fulfilled",
      ).length;
      const failed = results.filter((r) => r.status === "rejected").length;

      if (failed > 0 && succeeded === 0) {
        throw new Error(`Failed to revoke all ${failed} tokens`);
      }

      return { succeeded, failed, total: tokenIds.length };
    },
    onSuccess: () => {
      void queryClient.invalidateQueries(tokenKeys.userTokens());
    },
    onError: () => {
      void queryClient.invalidateQueries(tokenKeys.userTokens());
    },
  });
}