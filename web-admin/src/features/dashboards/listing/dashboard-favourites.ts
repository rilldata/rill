import { SvelteLocalStorage } from "@rilldata/web-common/lib/store-utils/svelte-local-storage.svelte.ts";

export function getDashboardFavouritesStore(org: string, project: string) {
  const key = `rill:app:${org}:${project}:dashboard:favourites`;
  return SvelteLocalStorage.createStringArrayStore(key);
}

export function getDashboardTagFavouritesStore(org: string, project: string) {
  const key = `rill:app:${org}:${project}:tag:favourites`;
  return SvelteLocalStorage.createStringArrayStore(key);
}

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
