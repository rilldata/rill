import {
  selectedMockUserJWT,
  selectedMockUserStore,
} from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
import type { MockUser } from "@rilldata/web-common/features/dashboards/granular-access-policies/useMockUsers";
import { runtimeServiceIssueDevJWT } from "@rilldata/web-common/runtime-client";
import { invalidateAllMetricsViews } from "@rilldata/web-common/runtime-client/invalidation";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { QueryClient } from "@tanstack/svelte-query";

export async function updateDevJWT(
  queryClient: QueryClient,
  client: RuntimeClient,
  mockUser: MockUser | null,
) {
  selectedMockUserStore.set(mockUser);

  if (mockUser === null) {
    selectedMockUserJWT.set(null);
    client.updateJwt(undefined, "user");
  } else {
    try {
      const { name, email, groups, admin, ...customAttributes } = mockUser;

      const { jwt } = await runtimeServiceIssueDevJWT(client, {
        email,
        name: name || "Mock User",
        admin: !!admin,
        groups: groups || [],
        attributes: customAttributes,
      });

      if (!jwt) throw new Error("No JWT returned");

      selectedMockUserJWT.set(jwt);
      client.updateJwt(jwt, "mock");
    } catch {
      // no-op
    }
  }

  return invalidateAllMetricsViews(queryClient, client.instanceId);
}
