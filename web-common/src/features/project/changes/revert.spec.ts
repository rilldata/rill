import { describe, it, expect, vi, beforeEach } from "vitest";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

// Mock only the raw revert RPC; keep the real query-key helpers so we assert against the actual
// key shape the app uses.
const gitRevertMock = vi.fn();
vi.mock("@rilldata/web-common/runtime-client", async (importActual) => {
  const actual =
    await importActual<typeof import("@rilldata/web-common/runtime-client")>();
  return {
    ...actual,
    runtimeServiceGitRevert: (...args: unknown[]) => gitRevertMock(...args),
  };
});

// Imported after the mock is registered so revert.ts picks up the mocked RPC.
const { revertFiles } = await import("./revert");

const client = { instanceId: "inst-1" } as RuntimeClient;

describe("revertFiles", () => {
  beforeEach(() => {
    gitRevertMock.mockReset();
  });

  it("calls the revert RPC with the branch and paths and returns the reverted count", async () => {
    gitRevertMock.mockResolvedValue({ revertedPaths: ["a.yaml", "b.sql"] });
    const invalidate = vi
      .spyOn(queryClient, "invalidateQueries")
      .mockResolvedValue(undefined);

    const count = await revertFiles(client, "main", ["a.yaml", "b.sql"]);

    expect(count).toBe(2);
    expect(gitRevertMock).toHaveBeenCalledWith(client, {
      remoteBranch: "main",
      paths: ["a.yaml", "b.sql"],
    });

    // Both GitDiff and GitStatus are invalidated by their shared key prefix so every param variant
    // (list vs full-diff, current branch vs primary branch) refetches.
    const keys = invalidate.mock.calls.map((c) => c[0]?.queryKey);
    expect(keys).toContainEqual(["RuntimeService", "gitDiff", "inst-1"]);
    expect(keys).toContainEqual(["RuntimeService", "gitStatus", "inst-1"]);

    invalidate.mockRestore();
  });

  it("returns 0 when the response has no reverted paths", async () => {
    gitRevertMock.mockResolvedValue({});
    vi.spyOn(queryClient, "invalidateQueries").mockResolvedValue(undefined);

    const count = await revertFiles(client, undefined, []);
    expect(count).toBe(0);
  });
});
