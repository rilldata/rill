import { SvelteLocalStorage } from "@rilldata/web-common/lib/store-utils/svelte-local-storage.svelte.ts";
import type { SortOption } from "@rilldata/web-common/components/table-toolbar";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { resourceKey } from "@rilldata/web-common/features/resources/overview-utils.ts";

/**
 * Favourited dashboards, keyed by `resourceKey(kind, name)`.
 * Entries written before the kind was included are bare lowercase names;
 * see migrateLegacyDashboardFavourites.
 */
export function getDashboardFavouritesStore(org: string, project: string) {
  const key = `rill:app:${org}:${project}:dashboard:favourites`;
  return SvelteLocalStorage.createStringArrayStore(key);
}

/**
 * Rewrites legacy favourites (bare lowercase names) to `kind/name` keys.
 * A legacy name matches every dashboard with that name, whatever its kind;
 * names with no matching dashboard are dropped.
 * Returns undefined when there is nothing to migrate.
 */
export function migrateLegacyDashboardFavourites(
  favourites: string[],
  dashboards: V1Resource[],
): string[] | undefined {
  if (!favourites.some((f) => !f.includes("/"))) return undefined;
  const migrated = favourites.flatMap((f) => {
    if (f.includes("/")) return [f];
    return dashboards.flatMap((r) => {
      const kind = r.meta?.name?.kind;
      const name = r.meta?.name?.name;
      if (!kind || !name || name.toLowerCase() !== f) return [];
      return [resourceKey(kind, name)];
    });
  });
  return [...new Set(migrated)];
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

/**
 * Rewrites legacy recently-used entries (bare lowercase names) to `kind/name` keys.
 * A legacy name applies to every dashboard with that name, whatever its kind;
 * names with no matching dashboard are dropped.
 * When a key already exists, the newer timestamp wins.
 * Returns undefined when there is nothing to migrate.
 */
export function migrateLegacyRecentlyUsedDashboards(
  recentlyUsed: Record<string, number>,
  dashboards: V1Resource[],
): Record<string, number> | undefined {
  const legacyNames = Object.keys(recentlyUsed).filter((k) => !k.includes("/"));
  if (legacyNames.length === 0) return undefined;
  const migrated: Record<string, number> = {};
  for (const [key, ts] of Object.entries(recentlyUsed)) {
    if (key.includes("/")) migrated[key] = Math.max(migrated[key] ?? 0, ts);
  }
  for (const name of legacyNames) {
    for (const r of dashboards) {
      const kind = r.meta?.name?.kind;
      const resName = r.meta?.name?.name;
      if (!kind || !resName || resName.toLowerCase() !== name) continue;
      const key = resourceKey(kind, resName);
      migrated[key] = Math.max(migrated[key] ?? 0, recentlyUsed[name]);
    }
  }
  return migrated;
}

/**
 * Last-opened timestamps per dashboard, keyed by `resourceKey(kind, name)`.
 * Entries written before the kind was included are bare lowercase names;
 * see migrateLegacyRecentlyUsedDashboards.
 */
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

  public update(kind: string, name: string) {
    this.recentlyUsed.setter({
      ...this.recentlyUsed.value,
      [resourceKey(kind, name)]: Date.now(),
    });
  }
}
