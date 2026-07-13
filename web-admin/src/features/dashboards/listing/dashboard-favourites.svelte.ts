import { SvelteLocalStorage } from "@rilldata/web-common/lib/store-utils/svelte-local-storage.svelte.ts";
import { RecordRuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";
import { UrlParamsState } from "@rilldata/web-common/lib/store-utils/url-params-state.svelte.ts";
import type { SortOption } from "@rilldata/web-common/components/table-toolbar";

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
    value: "lastUsed",
    label: "Last Used",
  },
  {
    value: "name",
    label: "Name",
  },
];

export class DashboardTableSort {
  public value: Record<string, boolean>;
  public getter: () => Record<string, boolean>;
  public setter: (newValue: Record<string, boolean>) => void;
  public set: (key: string, value: boolean) => void;

  private store: RecordRuneStore<boolean>;

  public constructor(
    defaultValue: Record<string, boolean> = {
      lastUsed: true,
      name: false,
    },
  ) {
    this.store = new RecordRuneStore<boolean>(
      new UrlParamsState(
        "sort",
        (value: Record<string, boolean>) => {
          const values = Object.entries(value).map(
            ([key, order]) => `${key}:${order}`,
          );
          return values.length ? values.join(",") : null;
        },
        (value) => {
          const values =
            value?.split(",").map((kv) => {
              const [k, v] = kv.split(":");
              return [k, v === "true"] as [string, boolean];
            }) ?? [];
          return Object.fromEntries(values);
        },
        defaultValue ?? {},
      ),
    );
    this.value = $derived(this.store.value);
    this.getter = this.store.getter;
    this.setter = this.store.setter;
    this.set = this.store.set;
  }
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
