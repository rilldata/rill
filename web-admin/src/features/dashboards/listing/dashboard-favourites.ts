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
 * Rewrites legacy favourites (bare lowercase names) to `kind/name` keys.
 * A legacy name matches every dashboard with that name, whatever its kind.
 * A name with no matching dashboard is kept as is: it is ignored when pinning
 * and can still be migrated once the dashboard is listed again.
 * Returns undefined when nothing changed.
 */
export function migrateLegacyDashboardFavourites(
  favourites: string[],
  dashboards: V1Resource[],
): string[] | undefined {
  const migrated = [
    ...new Set(
      favourites.flatMap((f) => {
        if (f.includes("/")) return [f];
        const keys = legacyNameToKeys(f, dashboards);
        return keys.length ? keys : [f];
      }),
    ),
  ];
  const unchanged =
    migrated.length === favourites.length &&
    migrated.every((f, i) => f === favourites[i]);
  return unchanged ? undefined : migrated;
}

/**
 * Rewrites legacy recently-used entries (bare lowercase names) to `kind/name` keys.
 * A legacy name applies to every dashboard with that name, whatever its kind;
 * a name with no matching dashboard is kept as is.
 * When a key already exists, the newer timestamp wins.
 * Returns undefined when nothing changed.
 */
export function migrateLegacyRecentlyUsedDashboards(
  recentlyUsed: Record<string, number>,
  dashboards: V1Resource[],
): Record<string, number> | undefined {
  const migrated: Record<string, number> = {};
  let changed = false;
  for (const [key, ts] of Object.entries(recentlyUsed)) {
    const keys = key.includes("/") ? [] : legacyNameToKeys(key, dashboards);
    if (keys.length === 0) {
      migrated[key] = Math.max(migrated[key] ?? 0, ts);
      continue;
    }
    changed = true;
    for (const k of keys) migrated[k] = Math.max(migrated[k] ?? 0, ts);
  }
  return changed ? migrated : undefined;
}

/**
 * Migrates both per-project dashboard stores in one go.
 * Called once the dashboards list is loaded, from a component mounted on every project page.
 */
export function migrateLegacyDashboardStores(
  org: string,
  project: string,
  dashboards: V1Resource[],
) {
  const favourites = getDashboardFavouritesStore(org, project);
  const migratedFavourites = migrateLegacyDashboardFavourites(
    favourites.value,
    dashboards,
  );
  if (migratedFavourites) favourites.setter(migratedFavourites);

  const recentlyUsed = new RecentlyUsedDashboards(org, project).recentlyUsed;
  const migratedRecentlyUsed = migrateLegacyRecentlyUsedDashboards(
    recentlyUsed.value,
    dashboards,
  );
  if (migratedRecentlyUsed) recentlyUsed.setter(migratedRecentlyUsed);
}

function legacyNameToKeys(name: string, dashboards: V1Resource[]): string[] {
  return dashboards.flatMap((r) => {
    const kind = r.meta?.name?.kind;
    const resName = r.meta?.name?.name;
    if (!kind || !resName || resName.toLowerCase() !== name) return [];
    return [resourceKey(kind, resName)];
  });
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
