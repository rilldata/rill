import {
  selectedMockUserJWT,
  selectedMockUserStore,
} from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
import type { MockUser } from "@rilldata/web-common/features/dashboards/granular-access-policies/useMockUsers";
import { runtimeServiceIssueDevJWT } from "@rilldata/web-common/runtime-client";
import { invalidateAllMetricsViews } from "@rilldata/web-common/runtime-client/invalidation";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export async function updateDevJWT(
  queryClient: QueryClient,
  mockUser: MockUser | null,
) {
  selectedMockUserStore.set(mockUser);

  if (mockUser === null) {
    selectedMockUserJWT.set(null);
    runtime.update((runtimeState) => {
      runtimeState.jwt = undefined;
      return runtimeState;
    });
  } else {
    try {
      const { jwt } = await runtimeServiceIssueDevJWT({
        name: mockUser?.name ? mockUser.name : "Mock User",
        email: mockUser?.email,
        groups: mockUser?.groups ? mockUser.groups : [],
        admin: !!mockUser?.admin,
      });

      if (!jwt) throw new Error("No JWT returned");

      selectedMockUserJWT.set(jwt);

      runtime.update((runtimeState) => {
        runtimeState.jwt = {
          token: jwt,
          receivedAt: Date.now(),
        };
        return runtimeState;
      });
    } catch (e) {
      // no-op
    }
  }

  return invalidateAllMetricsViews(queryClient, get(runtime).instanceId);
}
