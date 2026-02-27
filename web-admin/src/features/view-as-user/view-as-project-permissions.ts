import {
  createAdminServiceGetDeploymentCredentials,
  type V1GetDeploymentCredentialsResponse,
  type V1GetProjectResponse,
  type V1ProjectPermissions,
} from "@rilldata/web-admin/client";
import { createAdminServiceGetProjectWithBearerToken } from "@rilldata/web-admin/features/public-urls/get-project-with-bearer-token";
import type { CreateQueryResult } from "@tanstack/svelte-query";

/**
 * Creates a query for fetching deployment credentials for a mocked/simulated user.
 *
 * This is the first step in the View As query chain.
 * TanStack Query deduplicates by query key, so multiple consumers get instant cache hits.
 */
export function createViewAsCredentialsQuery(
  org: string,
  project: string,
  mockedUserId: string | undefined,
): CreateQueryResult<V1GetDeploymentCredentialsResponse> {
  return createAdminServiceGetDeploymentCredentials(
    org,
    project,
    { userId: mockedUserId },
    {
      query: {
        enabled: !!mockedUserId,
      },
    },
  );
}

/**
 * Creates a query for fetching project data using a mocked user's JWT.
 *
 * This is the second step in the View As query chain.
 * It fetches the project using the mocked user's access token to get their permissions.
 */
export function createViewAsProjectQuery(
  org: string,
  project: string,
  accessToken: string | undefined,
): CreateQueryResult<V1GetProjectResponse> {
  return createAdminServiceGetProjectWithBearerToken(
    org,
    project,
    accessToken ?? "",
    undefined,
    {
      query: {
        enabled: !!accessToken,
      },
    },
  );
}

/**
 * Computes effective project permissions based on View As state.
 *
 * When View As is active (mockedUserId is set), returns the mocked user's permissions.
 * Otherwise, returns the actual user's permissions.
 */
export function computeEffectivePermissions(
  mockedUserId: string | undefined,
  mockedPermissions: V1ProjectPermissions | undefined,
  actualPermissions: V1ProjectPermissions | undefined,
): V1ProjectPermissions | undefined {
  return mockedUserId && mockedPermissions
    ? mockedPermissions
    : actualPermissions;
}
