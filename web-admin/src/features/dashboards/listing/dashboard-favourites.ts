import { SvelteLocalStorage } from "@rilldata/web-common/lib/store-utils/svelte-local-storage.svelte.ts";
import type { SortOption } from "@rilldata/web-common/components/table-toolbar";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

export function getDashboardFavouritesStore(org: string, project: string) {
  const key = `rill:app:${org}:${project}:dashboard:favourites`;
  return SvelteLocalStorage.createStringArrayStore(key);
}

export function getDashboardTagFavouritesStore(org: string, project: string) {
  const key = `rill:app:${org}:${project}:tag:favourites`;
  return SvelteLocalStorage.createStringArrayStore(key);
}

/**
 * Sorts items so that favourites come first, in the order they were favourited,
 * with everything else following in its original (stable) order.
 */
export function sortByFavourites<T>(
  items: T[],
  favourites: string[],
  key: (item: T) => string,
): T[] {
  return [...items].sort((a, b) => {
    const aIndex = favourites.indexOf(key(a));
    const bIndex = favourites.indexOf(key(b));
    return (
      (aIndex === -1 ? favourites.length : aIndex) -
      (bIndex === -1 ? favourites.length : bIndex)
    );
  });
}

export const DashboardTableSortOptions: SortOption[] = [
  {
    value: "last_used_desc",
    get label() {
      return m.dashboard_sort_last_used();
    },
    sort: {
      id: "lastUsed",
      desc: true,
    },
  },
  {
    value: "name_asc",
    get label() {
      return m.dashboard_sort_name();
    },
    sort: {
      id: "name",
      desc: false,
    },
  },
];

export class RecentlyUsedDashboards {
  public readonly recentlyUsed: SvelteLocalStorage<
    Record<string, number>,
    Record<string, number>
  >;

  public constructor(
    public org: string,
    public project: string,
  ) {
    this.recentlyUsed = SvelteLocalStorage.getInstance(
      `rill:app:${org}:${project}:dashboard:recentlyUsed`,
      (value: Record<string, number>) => JSON.stringify(value),
      (value) => (value ? JSON.parse(value) : {}),
      {} as Record<string, number>,
    );
  }

  public update(dashboardName: string) {
    this.recentlyUsed.setter({
      ...this.recentlyUsed.value,
      [dashboardName]: Date.now(),
    });
  }
}
