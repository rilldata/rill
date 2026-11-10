import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { readable } from "svelte/store";
import BookmarksTable from "./BookmarksTable.svelte";

const navigation = vi.hoisted(() => ({
  callbacks: [] as ((event: { to: { url: URL } | null }) => void)[],
  goto: vi.fn(),
}));
vi.mock("$app/state", () => ({
  page: {
    params: { organization: "e2e", project: "openrtb" },
    url: new URL("http://localhost/e2e/openrtb/-/bookmarks?q=manager"),
    state: {},
  },
}));
vi.mock("$app/navigation", () => ({
  beforeNavigate: (callback: (event: { to: { url: URL } | null }) => void) =>
    navigation.callbacks.push(callback),
  goto: navigation.goto,
}));
vi.mock("@rilldata/web-admin/client", async (original) => ({
  ...(await original<typeof import("@rilldata/web-admin/client")>()),
  createAdminServiceGetCurrentUser: () =>
    readable({ data: { user: { id: "user" } }, isPending: false }),
  createAdminServiceRemoveBookmark: () =>
    readable({ mutateAsync: vi.fn(), isPending: false }),
}));
vi.mock("@tanstack/svelte-query", async (original) => ({
  ...(await original<typeof import("@tanstack/svelte-query")>()),
  createQuery: () =>
    readable({ data: { bookmarks: [] }, isPending: false, isError: false }),
}));
vi.mock(
  "@rilldata/web-admin/features/bookmarks/selectors.ts",
  async (original) => ({
    ...(await original<
      typeof import("@rilldata/web-admin/features/bookmarks/selectors.ts")
    >()),
    getProjectBookmarksQueryOptions: () => ({ queryKey: ["bookmarks"] }),
    invalidateBookmarkQueries: vi.fn(),
  }),
);
vi.mock("@rilldata/web-admin/features/dashboards/listing/selectors.ts", () => ({
  useDashboards: () => readable({ data: [] }),
}));
vi.mock(
  "@rilldata/web-admin/features/projects/selectors.ts",
  async (original) => ({
    ...(await original<
      typeof import("@rilldata/web-admin/features/projects/selectors.ts")
    >()),
    getProjectPermissions: () => readable({ data: { manageBookmarks: true } }),
  }),
);
vi.mock("@rilldata/web-common/runtime-client/v2", async (original) => ({
  ...(await original<
    typeof import("@rilldata/web-common/runtime-client/v2")
  >()),
  useRuntimeClient: () => ({ instanceId: "default" }),
}));

beforeEach(() => {
  vi.useFakeTimers();
  navigation.callbacks.length = 0;
  navigation.goto.mockClear();
  vi.stubGlobal(
    "ResizeObserver",
    class {
      observe() {}
      unobserve() {}
      disconnect() {}
    },
  );
});
afterEach(async () => {
  cleanup();
  await vi.advanceTimersByTimeAsync(50);
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

async function clearSearch() {
  await fireEvent.input(screen.getByPlaceholderText("Search..."), {
    target: { value: "" },
  });
  await vi.advanceTimersByTimeAsync(400);
}

describe("bookmark search navigation lifecycle", () => {
  it.each([
    "http://localhost/e2e/openrtb/explore/auction_explore_bookmarks",
    null,
  ])(
    "does not issue a late manager navigation when leaving for %s",
    async (url) => {
      render(BookmarksTable);
      await clearSearch();
      // Drive the callback actually registered by the component, before the 500 ms debounce fires.
      navigation.callbacks.forEach((callback) =>
        callback({ to: url ? { url: new URL(url) } : null }),
      );
      await vi.advanceTimersByTimeAsync(200);
      expect(navigation.callbacks).toHaveLength(1);
      expect(navigation.goto).not.toHaveBeenCalled();
    },
  );

  it("keeps pending search writes during same-path query navigation", async () => {
    render(BookmarksTable);
    await clearSearch();
    navigation.callbacks.forEach((callback) =>
      callback({
        to: {
          url: new URL("http://localhost/e2e/openrtb/-/bookmarks?sort=name"),
        },
      }),
    );
    await vi.advanceTimersByTimeAsync(200);
    expect(navigation.goto).toHaveBeenCalledTimes(1);
    const url = navigation.goto.mock.calls[0][0] as URL;
    expect(url.pathname).toBe("/e2e/openrtb/-/bookmarks");
    expect(url.searchParams.has("q")).toBe(false);
  });

  it("discards a pending search when the table is unmounted", async () => {
    const { unmount } = render(BookmarksTable);
    await clearSearch();
    unmount();
    await vi.advanceTimersByTimeAsync(200);
    expect(navigation.goto).not.toHaveBeenCalled();
  });
});
