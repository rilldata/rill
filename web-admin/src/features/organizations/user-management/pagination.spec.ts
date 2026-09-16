import { InfiniteQueryObserver, QueryClient } from "@tanstack/query-core";
import { afterEach, describe, expect, it, vi } from "vitest";
import { loadNextInvitePageForFilter } from "./pagination";

const clients: QueryClient[] = [];

function createObserver(total: number) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  clients.push(client);
  const queryFn = vi.fn(({ pageParam }: { pageParam: number }) =>
    Promise.resolve({
      emails: Array.from(
        { length: Math.min(50, total - pageParam) },
        (_, i) => `user${pageParam + i}@example.com`,
      ),
      nextPageToken: pageParam + 50 < total ? pageParam + 50 : undefined,
    }),
  );
  const observer = new InfiniteQueryObserver(client, {
    queryKey: ["users"],
    queryFn,
    initialPageParam: 0,
    getNextPageParam: (page) => page.nextPageToken,
  });
  return { observer, queryFn };
}

afterEach(() => {
  clients.splice(0).forEach((client) => client.clear());
});

describe("loading invitation pages for organization user filters", () => {
  it.each([
    ["invitations", 168],
    ["pending invitations", 101],
    ["guest invitations", 151],
  ])("finds %s beyond the first page without scrolling", async (_, total) => {
    const { observer, queryFn } = createObserver(total);
    let hasActiveFilters = false;
    const unsubscribe = observer.subscribe((query) => {
      loadNextInvitePageForFilter(query, hasActiveFilters);
    });

    await vi.waitFor(() =>
      expect(observer.getCurrentResult().isSuccess).toBe(true),
    );
    expect(queryFn).toHaveBeenCalledTimes(1);
    const matchingEmails = () =>
      observer
        .getCurrentResult()
        .data?.pages.flatMap((page) =>
          page.emails.filter(
            (email) => email === `user${total - 1}@example.com`,
          ),
        );
    expect(matchingEmails()).toEqual([]);

    hasActiveFilters = true;
    loadNextInvitePageForFilter(observer.getCurrentResult(), hasActiveFilters);

    await vi.waitFor(() =>
      expect(observer.getCurrentResult().hasNextPage).toBe(false),
    );
    expect(matchingEmails()).toEqual([`user${total - 1}@example.com`]);
    expect(queryFn).toHaveBeenCalledTimes(Math.ceil(total / 50));
    unsubscribe();
  });

  it("stops loading remaining pages when filters are cleared", async () => {
    const { observer, queryFn } = createObserver(168);
    let hasActiveFilters = false;
    const unsubscribe = observer.subscribe((query) => {
      loadNextInvitePageForFilter(query, hasActiveFilters);
    });
    await vi.waitFor(() =>
      expect(observer.getCurrentResult().isSuccess).toBe(true),
    );

    hasActiveFilters = true;
    loadNextInvitePageForFilter(observer.getCurrentResult(), hasActiveFilters);
    hasActiveFilters = false;

    await vi.waitFor(() =>
      expect(observer.getCurrentResult().isFetching).toBe(false),
    );
    expect(queryFn).toHaveBeenCalledTimes(2);
    expect(observer.getCurrentResult().hasNextPage).toBe(true);
    unsubscribe();
  });

  it("does not repeatedly fetch a failing page", async () => {
    const { observer, queryFn } = createObserver(168);
    let hasActiveFilters = false;
    const unsubscribe = observer.subscribe((query) => {
      loadNextInvitePageForFilter(query, hasActiveFilters);
    });
    await vi.waitFor(() =>
      expect(observer.getCurrentResult().isSuccess).toBe(true),
    );

    queryFn.mockRejectedValueOnce(new Error("Network error"));
    hasActiveFilters = true;
    loadNextInvitePageForFilter(observer.getCurrentResult(), hasActiveFilters);

    await vi.waitFor(() =>
      expect(observer.getCurrentResult().isError).toBe(true),
    );
    expect(queryFn).toHaveBeenCalledTimes(2);
    unsubscribe();
  });
});
