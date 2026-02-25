import {
  selectedMockUserJWT,
  selectedMockUserStore,
} from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
import type { MockUser } from "@rilldata/web-common/features/dashboards/granular-access-policies/useMockUsers";
import { invalidateAllMetricsViews } from "@rilldata/web-common/runtime-client/invalidation";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { runtimeServiceIssueDevJWT } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import type { QueryClient } from "@tanstack/svelte-query";

export async function updateDevJWT(
  queryClient: QueryClient,
  instanceId: string,
  mockUser: MockUser | null,
  runtimeClient: RuntimeClient,
) {
  selectedMockUserStore.set(mockUser);

  if (mockUser === null) {
    selectedMockUserJWT.set(null);
    runtimeClient.updateJwt(undefined, "user");
  } else {
    try {
      const { name, email, groups, admin, ...customAttributes } = mockUser;

      const { jwt } = await runtimeServiceIssueDevJWT(runtimeClient, {
        email,
        name: name || "Mock User",
        admin: !!admin,
        groups: groups || [],
        attributes: customAttributes,
      });

      if (!jwt) throw new Error("No JWT returned");

      selectedMockUserJWT.set(jwt);
      runtimeClient.updateJwt(jwt, "mock");
    } catch {
      // no-op
    }
  }

  return invalidateAllMetricsViews(queryClient, instanceId);
}
