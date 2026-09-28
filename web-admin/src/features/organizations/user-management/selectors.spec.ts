import { InfiniteQueryObserver, QueryClient } from "@tanstack/query-core";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { httpClient } from "@rilldata/web-admin/client/http-client";
import type { V1ListOrganizationMemberUsersResponse } from "@rilldata/web-admin/client";
import { getOrgUserMembersQueryOptions } from "./selectors";
import { invalidateOrgMemberUsers } from "./utils";

vi.mock("@rilldata/web-admin/client/http-client", () => ({
  httpClient: vi.fn(),
}));

const request = vi.mocked(httpClient);
const filters = { organization: "acme", guestOnly: false };
let client: QueryClient;

beforeEach(() => {
  request.mockReset();
  client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
});
afterEach(() => client.clear());

describe("organization member search", () => {
  it("searches all members on the server and only loads more matches on demand", async () => {
    request.mockResolvedValueOnce({
      members: [{ userEmail: "user1999@example.com" }],
      nextPageToken: "matching-page-2",
    });
    const observer = new InfiniteQueryObserver(
      client,
      getOrgUserMembersQueryOptions({
        ...filters,
        searchText: "1999",
        role: "viewer",
      }),
    );
    const unsubscribe = observer.subscribe(() => {});
    await vi.waitFor(() =>
      expect(observer.getCurrentResult().isSuccess).toBe(true),
    );
    expect(request).toHaveBeenCalledTimes(1);
    expect(request).toHaveBeenLastCalledWith(
      expect.objectContaining({
        params: expect.objectContaining({
          searchPattern: "%1999%",
          role: "viewer",
          pageSize: 50,
        }),
      }),
    );
    expect(
      observer.getCurrentResult().data?.pages[0].members?.[0].userEmail,
    ).toBe("user1999@example.com");

    request.mockResolvedValueOnce({ members: [], nextPageToken: "" });
    await observer.fetchNextPage();
    expect(request).toHaveBeenLastCalledWith(
      expect.objectContaining({
        params: expect.objectContaining({
          pageToken: "matching-page-2",
          searchPattern: "%1999%",
          role: "viewer",
        }),
      }),
    );
    unsubscribe();
  });

  it("keeps previous rows while a new search/role loads, then replaces them", async () => {
    const oldRow = { userEmail: "old@example.com", roleName: "viewer" };
    const newRow = { userEmail: "new@example.com", roleName: "editor" };
    request.mockResolvedValueOnce({
      members: [oldRow],
      nextPageToken: "old-page-2",
    });
    const observer = new InfiniteQueryObserver(
      client,
      getOrgUserMembersQueryOptions(filters),
    );
    const unsubscribe = observer.subscribe(() => {});
    await vi.waitFor(() =>
      expect(observer.getCurrentResult().isSuccess).toBe(true),
    );

    let resolveSearch!: (value: V1ListOrganizationMemberUsersResponse) => void;
    request.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveSearch = resolve;
      }),
    );
    observer.setOptions(
      getOrgUserMembersQueryOptions({
        ...filters,
        searchText: "new",
        role: "editor",
      }),
    );
    expect(observer.getCurrentResult().isPlaceholderData).toBe(true);
    expect(observer.getCurrentResult().isFetching).toBe(true);
    expect(observer.getCurrentResult().data?.pages[0].members).toEqual([
      oldRow,
    ]);
    expect(request).toHaveBeenLastCalledWith(
      expect.objectContaining({
        params: expect.objectContaining({
          searchPattern: "%new%",
          role: "editor",
          pageToken: undefined,
        }),
      }),
    );

    resolveSearch({ members: [newRow], nextPageToken: "" });
    await vi.waitFor(() =>
      expect(observer.getCurrentResult().isPlaceholderData).toBe(false),
    );
    expect(observer.getCurrentResult().data?.pages[0].members).toEqual([
      newRow,
    ]);

    request.mockResolvedValueOnce({ members: [newRow], nextPageToken: "" });
    await invalidateOrgMemberUsers(client, "acme");
    // Only the active filtered page is refetched after a mutation.
    expect(request).toHaveBeenCalledTimes(3);
    unsubscribe();
  });

  it("does not retain another organization's rows", async () => {
    request.mockResolvedValueOnce({
      members: [{ userEmail: "private@acme.com" }],
    });
    const observer = new InfiniteQueryObserver(
      client,
      getOrgUserMembersQueryOptions(filters),
    );
    const unsubscribe = observer.subscribe(() => {});
    await vi.waitFor(() =>
      expect(observer.getCurrentResult().isSuccess).toBe(true),
    );
    request.mockReturnValueOnce(new Promise(() => {}));
    observer.setOptions(
      getOrgUserMembersQueryOptions({ ...filters, organization: "other" }),
    );
    expect(observer.getCurrentResult().data).toBeUndefined();
    unsubscribe();
  });

  it("escapes literal ILIKE characters and keeps the guest role", async () => {
    request.mockResolvedValue({ members: [] });
    await client.fetchInfiniteQuery(
      getOrgUserMembersQueryOptions({
        ...filters,
        guestOnly: true,
        searchText: "a_b%\\c",
        role: "editor",
      }),
    );
    expect(request).toHaveBeenCalledWith(
      expect.objectContaining({
        params: expect.objectContaining({
          searchPattern: "%a\\_b\\%\\\\c%",
          role: "guest",
        }),
      }),
    );
  });

  it("does not fetch members for pending-only views", () => {
    const observer = new InfiniteQueryObserver(
      client,
      getOrgUserMembersQueryOptions({ ...filters, enabled: false }),
    );
    const unsubscribe = observer.subscribe(() => {});
    expect(request).not.toHaveBeenCalled();
    unsubscribe();
  });
});
