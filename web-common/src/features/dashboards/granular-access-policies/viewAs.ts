import type { QueryClient } from "@tanstack/svelte-query";
import { get, writable } from "svelte/store";
import {
  getRuntimeServiceGetCatalogEntryQueryKey,
  runtimeServiceIssueDevJWT,
} from "../../../runtime-client";
import { invalidateMetricsViewData } from "../../../runtime-client/invalidation";
import { runtime } from "../../../runtime-client/runtime-store";
import type { MockUser } from "./useMockUsers";

// Note: `null` means "viewing as self"
export const viewAsStore = writable<MockUser | null>(null);

export async function viewAs(
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
        admin: user?.admin,
      });
      console.log("resp", resp);

      // TODO: test this
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
  // Invalidate dashboard catalog entry
  await queryClient.invalidateQueries(
    getRuntimeServiceGetCatalogEntryQueryKey(instanceId, dashboardName)
  );

  // Invalidate dashboard queries
  invalidateMetricsViewData(queryClient, dashboardName, false);

  // Update store
  viewAsStore.set(user);
}
