import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import {
  getRuntimeServiceGetCatalogEntryQueryKey,
  runtimeServiceIssueDevJWT,
} from "../../../runtime-client";
import { invalidateMetricsViewData } from "../../../runtime-client/invalidation";
import { runtime } from "../../../runtime-client/runtime-store";
import { mockUserHasNoAccessStore, selectedMockUserStore } from "./stores";
import type { MockUser } from "./useMockUsers";

export async function viewAsMockUser(
  queryClient: QueryClient,
  dashboardName: string,
  user: MockUser | null
) {
  const instanceId = get(runtime).instanceId;

  if (user === null) {
    // Remove any Dev JWT from the runtime store
    runtime.set({
      ...get(runtime),
      jwt: null,
    });
  } else {
    try {
      // Issue Dev JWT
      const resp = await runtimeServiceIssueDevJWT({
        name: user?.name,
        email: user?.email,
        groups: user?.groups,
        admin: user?.admin ? true : false,
      });

      // Set the Dev JWT in the runtime store
      runtime.set({
        ...get(runtime),
        jwt: resp.jwt,
      });
    } catch (e) {
      console.error("Error issuing Dev JWT", e);

      // TODO: handle error
      // Is this where I should show a mini 404 page, or is that covered by invalidating the catalog entry?
    }
  }

  // Reset mockUserHasNoAccessStore
  mockUserHasNoAccessStore.set(false);

  // Invalidate dashboard catalog entry
  await queryClient.invalidateQueries(
    getRuntimeServiceGetCatalogEntryQueryKey(instanceId, dashboardName)
  );

  // Invalidate dashboard queries
  invalidateMetricsViewData(queryClient, dashboardName, false);

  // Update store
  selectedMockUserStore.set(user);
}
