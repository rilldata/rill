import { isAdminServerQuery } from "@rilldata/web-admin/client/utils";
import {
  clearAdminNetworkErrorState,
  handleAdminServerNetworkError,
  handleAdminServerQuerySuccess,
  recoverFromAdminNetworkError,
} from "@rilldata/web-admin/components/errors/admin-network-errors";
import { errorStore } from "@rilldata/web-admin/components/errors/error-store";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import type { Query, QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

type TestQueryClient = Pick<QueryClient, "refetchQueries">;

function makeQuery(queryKey: unknown[], data?: unknown): Query {
  return {
    queryKey,
    state: { data },
  } as unknown as Query;
}

function makeQueryClient(): TestQueryClient {
  return {
    refetchQueries: vi.fn(() => Promise.resolve()),
  } as unknown as TestQueryClient;
}

describe("admin network errors", () => {
  beforeEach(() => {
    errorStore.reset();
    vi.spyOn(eventBus, "emit");
  });

  afterEach(() => {
    clearAdminNetworkErrorState();
    vi.restoreAllMocks();
  });

  it("keeps cached data visible and shows a banner for admin network errors", () => {
    const queryClient = makeQueryClient();
    const handled = handleAdminServerNetworkError(
      new Error("Network Error"),
      makeQuery(["/v1/projects/org/proj"], { project: { name: "proj" } }),
      queryClient,
    );

    expect(handled).toBe(true);
    expect(get(errorStore).header).toBe("");
    expect(eventBus.emit).toHaveBeenCalledWith(
      "add-banner",
      expect.objectContaining({
        id: "admin-network",
        message: expect.objectContaining({
          message: expect.stringContaining("Showing cached data"),
          type: "warning",
        }),
      }),
    );
  });

  it("uses the full-page error only when there is no cached data", () => {
    const handled = handleAdminServerNetworkError(
      new Error("Network Error"),
      makeQuery(["/v1/projects/org/proj"]),
      makeQueryClient(),
    );

    expect(handled).toBe(true);
    expect(get(errorStore).header).toBe("Network Error");
    expect(eventBus.emit).not.toHaveBeenCalledWith(
      "add-banner",
      expect.anything(),
    );
  });

  it("ignores non-admin query network errors", () => {
    const handled = handleAdminServerNetworkError(
      new Error("Network Error"),
      makeQuery(["/v1/instances/runtime/api/query"], { rows: [] }),
      makeQueryClient(),
    );

    expect(handled).toBe(false);
    expect(get(errorStore).header).toBe("");
    expect(eventBus.emit).not.toHaveBeenCalledWith(
      "add-banner",
      expect.anything(),
    );
  });

  it("clears the banner after a successful admin query", () => {
    handleAdminServerNetworkError(
      new Error("Network Error"),
      makeQuery(["/v1/projects/org/proj"], { project: { name: "proj" } }),
      makeQueryClient(),
    );
    vi.mocked(eventBus.emit).mockClear();

    handleAdminServerQuerySuccess(makeQuery(["/v1/users/me"], { user: {} }));

    expect(eventBus.emit).toHaveBeenCalledWith(
      "remove-banner",
      "admin-network",
    );
  });

  it("refetches active admin queries during recovery", async () => {
    const queryClient = makeQueryClient();
    handleAdminServerNetworkError(
      new Error("Network Error"),
      makeQuery(["/v1/projects/org/proj"], { project: { name: "proj" } }),
      queryClient,
    );

    await recoverFromAdminNetworkError(queryClient);

    expect(queryClient.refetchQueries).toHaveBeenCalledWith({
      type: "active",
      predicate: isAdminServerQuery,
    });
  });

  it("treats non-string query keys as non-admin queries", () => {
    expect(isAdminServerQuery(makeQuery([{ path: "/v1/projects" }]))).toBe(
      false,
    );
  });
});
