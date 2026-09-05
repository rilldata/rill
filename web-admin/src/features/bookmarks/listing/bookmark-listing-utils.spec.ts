import type { V1Bookmark } from "@rilldata/web-admin/client";
import {
  buildBookmarkRows,
  filterBookmarkRows,
} from "@rilldata/web-admin/features/bookmarks/listing/bookmark-listing-utils.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

const dashboards: V1Resource[] = [
  {
    meta: { name: { kind: ResourceKind.Explore, name: "Auction_Explore" } },
    explore: { spec: { displayName: "Programmatic Ads Auction" } },
  },
  {
    meta: { name: { kind: ResourceKind.Canvas, name: "overview_canvas" } },
    canvas: { spec: { displayName: "" } },
  },
];

function bookmark(overrides: Partial<V1Bookmark>): V1Bookmark {
  return {
    id: "b1",
    displayName: "My bookmark",
    description: "",
    urlSearch: "?view=tdd&tr=P7D&f=domain IN ('a')",
    resourceKind: ResourceKind.Explore,
    resourceName: "auction_explore",
    userId: "u1",
    default: false,
    shared: false,
    ...overrides,
  };
}

function rowsFor(
  bookmarks: V1Bookmark[],
  manageBookmarks = false,
  recentlyUsed: Record<string, number> = {},
) {
  return buildBookmarkRows({
    bookmarks,
    dashboards,
    organization: "org",
    project: "proj",
    currentUserId: "u1",
    manageBookmarks,
    recentlyUsed,
  });
}

describe("buildBookmarkRows", () => {
  it("links explore bookmarks to the explore with the saved params", () => {
    const [row] = rowsFor([bookmark({})]);
    expect(row.href).toBe(
      "/org/proj/explore/auction_explore?view=tdd&tr=P7D&f=domain IN ('a')",
    );
    expect(row.dashboardKind).toBe(ResourceKind.Explore);
    expect(row.dashboardTitle).toBe("Programmatic Ads Auction");
    expect(row.category).toBe("personal");
    expect(row.filtersOnly).toBe(false);
    expect(row.isLegacy).toBe(false);
  });

  it("links canvas bookmarks to the canvas and falls back to the resource name as title", () => {
    const [row] = rowsFor([
      bookmark({
        resourceKind: ResourceKind.Canvas,
        resourceName: "overview_canvas",
        urlSearch: "tr=P1D",
      }),
    ]);
    expect(row.href).toBe("/org/proj/canvas/overview_canvas?tr=P1D");
    expect(row.dashboardKind).toBe(ResourceKind.Canvas);
    expect(row.dashboardTitle).toBe("overview_canvas");
  });

  it("uses the resource name when the dashboard is not in the project", () => {
    const [row] = rowsFor([bookmark({ resourceName: "deleted_explore" })]);
    expect(row.dashboardTitle).toBe("deleted_explore");
    expect(row.href).toBe(
      "/org/proj/explore/deleted_explore?view=tdd&tr=P7D&f=domain IN ('a')",
    );
  });

  it("does not link bookmarks for unknown resource kinds", () => {
    const [row] = rowsFor([
      bookmark({ resourceKind: "rill.runtime.v1.Report" }),
    ]);
    expect(row.href).toBeUndefined();
    expect(row.dashboardKind).toBeUndefined();
  });

  it("opens legacy bookmarks on the dashboard without params", () => {
    const [row] = rowsFor([bookmark({ urlSearch: "", data: "AAAA" })]);
    expect(row.isLegacy).toBe(true);
    expect(row.filtersOnly).toBe(false);
    expect(row.href).toBe("/org/proj/explore/auction_explore");
  });

  it("detects filters-only bookmarks", () => {
    const [row] = rowsFor([bookmark({ urlSearch: "?tr=P7D&grain=day" })]);
    expect(row.filtersOnly).toBe(true);
  });

  it("categorises home, managed and personal bookmarks", () => {
    const rows = rowsFor([
      bookmark({ id: "home", default: true, shared: true }),
      bookmark({ id: "managed", shared: true }),
      bookmark({ id: "personal" }),
    ]);
    expect(rows.map((r) => r.category)).toEqual([
      "home",
      "managed",
      "personal",
    ]);
  });

  it("lets owners manage personal bookmarks and admins manage shared ones", () => {
    const viewerRows = rowsFor([
      bookmark({ id: "mine" }),
      bookmark({ id: "theirs", userId: "u2" }),
      bookmark({ id: "managed", shared: true, userId: "u2" }),
      bookmark({ id: "home", default: true, shared: true, userId: "u2" }),
    ]);
    expect(viewerRows.map((r) => r.canManage)).toEqual([
      true,
      false,
      false,
      false,
    ]);

    const adminRows = rowsFor(
      [
        bookmark({ id: "managed", shared: true, userId: "u2" }),
        bookmark({ id: "home", default: true, shared: true, userId: "u2" }),
      ],
      true,
    );
    expect(adminRows.map((r) => r.canManage)).toEqual([true, true]);
  });
});

describe("buildBookmarkRows last used", () => {
  it("reads the last used time by bookmark id and defaults to zero", () => {
    const rows = rowsFor(
      [bookmark({ id: "opened" }), bookmark({ id: "never" })],
      false,
      { opened: 1700000000000 },
    );
    expect(rows.map((r) => r.lastUsed)).toEqual([1700000000000, 0]);
  });
});

describe("filterBookmarkRows", () => {
  const rows = rowsFor([
    bookmark({ id: "1", displayName: "Weekly review", description: "" }),
    bookmark({
      id: "2",
      displayName: "Top domains",
      description: "Daily check of top publishers",
    }),
    bookmark({
      id: "3",
      displayName: "Canvas view",
      resourceKind: ResourceKind.Canvas,
      resourceName: "overview_canvas",
    }),
  ]);

  it("returns all rows for an empty search", () => {
    expect(filterBookmarkRows(rows, "  ")).toHaveLength(3);
  });

  it("matches name, description and dashboard case-insensitively", () => {
    expect(
      filterBookmarkRows(rows, "WEEKLY").map((r) => r.bookmark.id),
    ).toEqual(["1"]);
    expect(
      filterBookmarkRows(rows, "publishers").map((r) => r.bookmark.id),
    ).toEqual(["2"]);
    expect(
      filterBookmarkRows(rows, "programmatic").map((r) => r.bookmark.id),
    ).toEqual(["1", "2"]);
    expect(
      filterBookmarkRows(rows, "overview_canvas").map((r) => r.bookmark.id),
    ).toEqual(["3"]);
  });
});
