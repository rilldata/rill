import { SvelteLocalStorage } from "@rilldata/web-common/lib/store-utils/svelte-local-storage.svelte.ts";

export function getDashboardFavouritesStore(org: string, project: string) {
  const key = `rill:app:${org}:${project}:dashboard:favourites`;
  return SvelteLocalStorage.createStringArrayStore(key);
}
