import type { Query, QueryClient } from "@tanstack/svelte-query";
import { describe, expect, it, vi } from "vitest";
import { invalidateAllMetricsViews } from "./invalidation";

const INSTANCE_ID = "instance-1";

function fakeQueryClient() {
  return {
    refetchQueries: vi.fn().mockResolvedValue(undefined),
    resetQueries: vi.fn().mockResolvedValue(undefined),
  } as unknown as QueryClient & {
    refetchQueries: ReturnType<typeof vi.fn>;
    resetQueries: ReturnType<typeof vi.fn>;
  };
}

function queryWithKey(queryKey: readonly unknown[]) {
  return { queryKey } as Query;
}

describe("invalidateAllMetricsViews", () => {
  it("refetches GetExplore metadata after the security context changes", async () => {
    const queryClient = fakeQueryClient();

    await invalidateAllMetricsViews(queryClient, INSTANCE_ID);

    const resourceRefetch = queryClient.refetchQueries.mock.calls[1]?.[0] as
      | { predicate?: (query: Query) => boolean }
      | undefined;
    const predicate = resourceRefetch?.predicate;

    expect(predicate).toBeDefined();
    expect(
      predicate?.(
        queryWithKey([
          "RuntimeService",
          "getExplore",
          INSTANCE_ID,
          { name: "campaigns" },
        ]),
      ),
    ).toBe(true);
    expect(
      predicate?.(
        queryWithKey([
          "RuntimeService",
          "getExplore",
          "another-instance",
          { name: "campaigns" },
        ]),
      ),
    ).toBe(false);
  });
});
