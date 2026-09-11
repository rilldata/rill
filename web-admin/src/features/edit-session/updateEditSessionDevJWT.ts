import {
  selectedMockUserJWT,
  selectedMockUserStore,
} from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
import type { MockUser } from "@rilldata/web-common/features/dashboards/granular-access-policies/useMockUsers";
import {
  invalidateAllMetricsViews,
  invalidateCanvasQueries,
} from "@rilldata/web-common/runtime-client/invalidation";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
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

    if (mockUser === null) {
      selectedMockUserJWT.set(null);
      client.updateJwt(editSessionJwt, "user");
    } else {
      try {
        const { name, email, groups, admin, ...customAttributes } = mockUser;

        const { accessToken } = await adminServiceGetDeployment(deploymentId, {
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
        });

        if (!accessToken) throw new Error("No JWT returned");

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
    return invalidateCanvasQueries(queryClient, client.instanceId);
  };
}
