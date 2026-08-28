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

export function createUpdateEditSessionDevJWT(
  deploymentId: string,
  editSessionJwt: string,
) {
  return async (
    queryClient: QueryClient,
    client: RuntimeClient,
    mockUser: MockUser | null,
  ) => {
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
            name: name || "Mock User",
            admin: !!admin,
            groups: groups || [],
            ...customAttributes,
          },
        });

        if (!accessToken) throw new Error("No JWT returned");

        selectedMockUserJWT.set(accessToken);
        client.updateJwt(accessToken, "mock");
      } catch {
        // no-op
      }
    }

    await invalidateAllMetricsViews(queryClient, client.instanceId);
    return invalidateCanvasQueries(queryClient, client.instanceId);
  };
}
