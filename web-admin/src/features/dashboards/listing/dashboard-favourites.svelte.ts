import { SvelteLocalStorage } from "@rilldata/web-common/lib/store-utils/svelte-local-storage.svelte.ts";
import { ArrayRuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";
import { UrlParamsState } from "@rilldata/web-common/lib/store-utils/url-params-state.svelte.ts";
import type { SortOption } from "@rilldata/web-common/components/table-toolbar";
import type { ColumnSort } from "tanstack-table-8-svelte-5";

export function getDashboardFavouritesStore(org: string, project: string) {
  const key = `rill:app:${org}:${project}:dashboard:favourites`;
  return SvelteLocalStorage.createStringArrayStore(key);
}

export function getDashboardTagFavouritesStore(org: string, project: string) {
  const key = `rill:app:${org}:${project}:tag:favourites`;
  return SvelteLocalStorage.createStringArrayStore(key);
}

export const DashboardTableSortOptions: SortOption[] = [
  {
    id: "lastUsed",
    label: "Last Used",
  },
  {
    id: "name",
    label: "Name",
  },
];

const DashboardTableSortDefault: ColumnSort[] = [
  { id: "lastUsed", desc: true },
];

export function createDashboardTableSortStore() {
  return new ArrayRuneStore<ColumnSort>(
    new UrlParamsState(
      "sort",
      (value: ColumnSort[]) => {
        const sortValues = value.map(({ id, desc }) => `${id}:${desc}`);
        return sortValues.length ? sortValues.join(",") : null;
      },
      (value) => {
        const values: ColumnSort[] =
          value?.split(",").map((kv) => {
            const [id, desc] = kv.split(":");
            return { id, desc: desc === "true" };
          }) ?? [];
        return values.length === 0 ? DashboardTableSortDefault : values;
      },
      DashboardTableSortDefault ?? [],
    ),
    (a, b) => a.id === b.id,
  );
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
