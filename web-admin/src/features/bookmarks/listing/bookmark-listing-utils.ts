import type { V1Bookmark } from "@rilldata/web-admin/client";
import { isFilterOnlyBookmark } from "@rilldata/web-admin/features/bookmarks/utils.ts";
import type { SortOption } from "@rilldata/web-common/components/table-toolbar";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

export type BookmarkCategory = "home" | "managed" | "personal";

export type BookmarkListRow = {
  bookmark: V1Bookmark;
  category: BookmarkCategory;
  // True when the bookmark stores only filter and time range params. Always false for legacy bookmarks, whose state is opaque here.
  filtersOnly: boolean;
  // Kind of the dashboard the bookmark is for. Undefined for kinds the manager cannot open.
  dashboardKind: ResourceKind.Explore | ResourceKind.Canvas | undefined;
  // Display name of the dashboard. Falls back to the resource name when the dashboard is not among the project's resources.
  dashboardTitle: string;
  // Opens the bookmark on its dashboard. Undefined when the dashboard kind is unknown.
  href: string | undefined;
  // The owner may manage personal bookmarks; managed and home bookmarks need the manage bookmarks permission.
  canManage: boolean;
  // Epoch milliseconds of the last time the bookmark was opened in this browser. Zero when never opened.
  lastUsed: number;
};

const DashboardSlugByKind: Partial<Record<string, "explore" | "canvas">> = {
  [ResourceKind.Explore]: "explore",
  [ResourceKind.Canvas]: "canvas",
};

export function buildBookmarkRows({
  bookmarks,
  dashboards,
  organization,
  project,
  currentUserId,
  manageBookmarks,
  recentlyUsed,
}: {
  bookmarks: V1Bookmark[];
  dashboards: V1Resource[];
  organization: string;
  project: string;
  currentUserId: string | undefined;
  manageBookmarks: boolean;
  // Bookmark id to epoch milliseconds, see RecentlyUsedBookmarks.
  recentlyUsed: Record<string, number>;
}): BookmarkListRow[] {
  const dashboardTitles = new Map<string, string>();
  for (const resource of dashboards) {
    const kind = resource.meta?.name?.kind;
    const name = resource.meta?.name?.name;
    if (!kind || !name) continue;
    const title =
      resource.explore?.spec?.displayName ?? resource.canvas?.spec?.displayName;
    dashboardTitles.set(dashboardKey(kind, name), title || name);
  }

  return bookmarks.map((bookmark) => {
    const kind = bookmark.resourceKind ?? "";
    const name = bookmark.resourceName ?? "";
    const slug = DashboardSlugByKind[kind];
    const urlSearch = bookmarkUrlSearch(bookmark);
    const category: BookmarkCategory = bookmark.default
      ? "home"
      : bookmark.shared
        ? "managed"
        : "personal";

    return {
      bookmark,
      category,
      filtersOnly: isFilterOnlyBookmark(new URLSearchParams(urlSearch)),
      dashboardKind: slug
        ? (kind as ResourceKind.Explore | ResourceKind.Canvas)
        : undefined,
      dashboardTitle: dashboardTitles.get(dashboardKey(kind, name)) ?? name,
      href: slug
        ? `/${organization}/${project}/${slug}/${name}${urlSearch}`
        : undefined,
      canManage:
        category === "personal"
          ? !!currentUserId && bookmark.userId === currentUserId
          : manageBookmarks,
      lastUsed: bookmark.id ? (recentlyUsed[bookmark.id] ?? 0) : 0,
    };
  });
}

export function filterBookmarkRows(
  rows: BookmarkListRow[],
  searchText: string,
): BookmarkListRow[] {
  const search = searchText.trim().toLowerCase();
  if (!search) return rows;
  return rows.filter((row) =>
    [
      row.bookmark.displayName,
      row.bookmark.description,
      row.dashboardTitle,
      row.bookmark.resourceName,
    ].some((text) => text?.toLowerCase().includes(search)),
  );
}

// Column ids refer to the hidden accessor columns declared in BookmarksTable.svelte.
export const BookmarkTableSortOptions: SortOption[] = [
  {
    value: "last_used_desc",
    get label() {
      return m.bookmark_sort_last_used();
    },
    sort: {
      id: "lastUsed",
      desc: true,
    },
  },
  {
    value: "updated_desc",
    get label() {
      return m.bookmark_sort_last_updated();
    },
    sort: {
      id: "updatedOn",
      desc: true,
    },
  },
  {
    value: "name_asc",
    get label() {
      return m.bookmark_sort_name();
    },
    sort: {
      id: "name",
      desc: false,
    },
  },
  {
    value: "dashboard_asc",
    get label() {
      return m.bookmark_sort_dashboard();
    },
    sort: {
      id: "dashboard",
      desc: false,
    },
  },
];

function dashboardKey(kind: string, name: string) {
  return `${kind}:${name.toLowerCase()}`;
}

// Search string that restores the bookmark's state on its dashboard, with a leading "?"; empty when the bookmark has no state.
// Legacy explore bookmarks store their state as proto data instead of url params;
// the explore page resolves that data through the "state" url param (see convertURLToExplorePreset).
// Canvas bookmarks never used that format, so their data is used as the search as-is, matching the canvas bookmark dropdown.
function bookmarkUrlSearch(bookmark: V1Bookmark) {
  let urlSearch = bookmark.urlSearch;
  if (!urlSearch && bookmark.data) {
    urlSearch =
      DashboardSlugByKind[bookmark.resourceKind ?? ""] === "explore"
        ? new URLSearchParams({
            [ExploreStateURLParams.LegacyProtoState]: bookmark.data,
          }).toString()
        : bookmark.data;
  }
  if (!urlSearch) return "";
  // Bookmarks created through the UI store the search with a leading "?"; tolerate ones that do not.
  return urlSearch.startsWith("?") ? urlSearch : `?${urlSearch}`;
}
