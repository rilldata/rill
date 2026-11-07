import {
  selectedMockUserJWT,
  selectedMockUserStore,
} from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
import type { MockUser } from "@rilldata/web-common/features/dashboards/granular-access-policies/useMockUsers";
import {
  invalidateAllMetricsViews,
  invalidateCanvasQueries,
  removeInactiveResourceQueries,
} from "@rilldata/web-common/runtime-client/invalidation";
import type {
  AuthContext,
  JwtRefresher,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import type { QueryClient } from "@tanstack/svelte-query";
import { adminServiceGetDeployment } from "@rilldata/web-admin/client";
import { getDomain } from "@rilldata/web-admin/features/projects/user-management/selectors.ts";
import { get } from "svelte/store";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { RUNTIME_ACCESS_TOKEN_DEFAULT_TTL } from "@rilldata/web-common/runtime-client/constants.ts";

export function createUpdateEditSessionDevJWT(
  deploymentId: string,
  editSessionJwt: string,
) {
  return async (
    queryClient: QueryClient,
    client: RuntimeClient,
    mockUser: MockUser | null,
  ) => {
    const prevMockUser = get(selectedMockUserStore);
    selectedMockUserStore.set(mockUser);

    // The edit layout derives the provider's credentials from the stores set here (see editSessionRuntimeAuth);
    // the client is updated directly as well so the invalidations below already run with the new token
    if (mockUser === null) {
      selectedMockUserJWT.set(null);
      client.updateJwt(editSessionJwt, "user");
    } else {
      try {
        const accessToken = await issueMockUserJwt(deploymentId, mockUser);
        selectedMockUserJWT.set(accessToken);
        client.updateJwt(accessToken, "mock");
      } catch {
        // Reset the user and make sure we are not in a errored state.
        selectedMockUserStore.set(prevMockUser);
        // There is no real action user can take so just show a notification for now.
        eventBus.emit("notification", {
          message: m.dashboard_view_as_error(),
          type: "error",
        });
      }
    }

    await invalidateAllMetricsViews(queryClient, client.instanceId);
    removeInactiveResourceQueries(queryClient, client.instanceId);
    return invalidateCanvasQueries(queryClient, client.instanceId);
  };
}

/**
 * Reissues the mock user's JWT when a request finds it expired, e.g. after a laptop sleep,
 * so the preview does not wait for a token that never comes. The mock token has the same TTL as the user's.
 */
export function createMockUserJwtRefresher(
  deploymentId: string,
  mockUser: MockUser,
): JwtRefresher {
  return async (signal) => {
    const accessToken = await issueMockUserJwt(deploymentId, mockUser, signal);
    selectedMockUserJWT.set(accessToken);
    return accessToken;
  };
}

/**
 * Picks the credentials the edit layout's RuntimeProvider runs with: the mock user's JWT and refresher while
 * View As is active, the user's otherwise. The provider must follow the View As selection: if it kept forwarding
 * the user's JWT, the next GetProject update (the 15-minute poll, a focus refetch, or an on-demand refresh that was
 * still in flight when View As started) would push the user's token back onto the client with the "user" context,
 * and the preview would show the user's data while View As stays selected.
 */
export function editSessionRuntimeAuth(opts: {
  userJwt: string | undefined;
  userRefreshJwt: JwtRefresher | undefined;
  mockUser: MockUser | null;
  mockJwt: string | null;
  mockRefreshJwt: JwtRefresher | undefined;
}): {
  jwt: string | undefined;
  authContext: AuthContext;
  refreshJwt: JwtRefresher | undefined;
} {
  if (opts.mockUser && opts.mockJwt) {
    return {
      jwt: opts.mockJwt,
      authContext: "mock",
      refreshJwt: opts.mockRefreshJwt,
    };
  }
  return {
    jwt: opts.userJwt,
    authContext: "user",
    refreshJwt: opts.userRefreshJwt,
  };
}

async function issueMockUserJwt(
  deploymentId: string,
  mockUser: MockUser,
  signal?: AbortSignal,
): Promise<string> {
  const { name, email, groups, admin, ...customAttributes } = mockUser;

  const { accessToken } = await adminServiceGetDeployment(
    deploymentId,
    {
      attributes: {
        email,
        ...(email ? { domain: getDomain(email) } : {}),
        name: name || "Mock User",
        admin: !!admin,
        groups: groups || [],
        ...customAttributes,
      },
      // Make sure to match TTL, GetDeployment defaults to 24-hours
      accessTokenTtlSeconds: RUNTIME_ACCESS_TOKEN_DEFAULT_TTL / 1000, // Seconds param vs milliseconds constant
    },
    signal,
  );

  if (!accessToken) throw new Error("No JWT returned");
  return accessToken;
}
