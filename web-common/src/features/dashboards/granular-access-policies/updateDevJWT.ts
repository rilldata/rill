import {
  selectedMockUserJWT,
  selectedMockUserStore,
} from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
import type { MockUser } from "@rilldata/web-common/features/dashboards/granular-access-policies/useMockUsers";
import { runtimeServiceIssueDevJWT } from "@rilldata/web-common/runtime-client";
import httpClient from "@rilldata/web-common/runtime-client/http-client";
import { invalidateAllMetricsViews } from "@rilldata/web-common/runtime-client/invalidation";
import type { QueryClient } from "@tanstack/svelte-query";

export async function updateDevJWT(
  queryClient: QueryClient,
  mockUser: MockUser | null,
) {
  selectedMockUserStore.set(mockUser);

  if (mockUser === null) {
    selectedMockUserJWT.set(null);

    await httpClient.updateJWT(undefined, undefined);
  } else {
    try {
      const { name, email, groups, admin, ...customAttributes } = mockUser;

      const { jwt } = await runtimeServiceIssueDevJWT({
        email,
        name: name || "Mock User",
        admin: !!admin,
        groups: groups || [],
        attributes: customAttributes,
      });

      if (!jwt) throw new Error("No JWT returned");

      selectedMockUserJWT.set(jwt);

      await httpClient.updateJWT(jwt, "mock");
    } catch {
      // no-op
    }
  }

  return invalidateAllMetricsViews(queryClient, httpClient.getInstanceId());
}
