import { SvelteLocalStorage } from "@rilldata/web-common/lib/store-utils/svelte-local-storage.svelte.ts";

/**
 * Tracks when each bookmark was last opened, per project, in local storage.
 * Mirrors RecentlyUsedDashboards: usage is a per-browser signal, not synced to the server.
 */
export class RecentlyUsedBookmarks {
  // Bookmark id to epoch milliseconds of the last time it was opened.
  public readonly recentlyUsed: SvelteLocalStorage<
    Record<string, number>,
    Record<string, number>
  >;

  public constructor(org: string, project: string) {
    this.recentlyUsed = SvelteLocalStorage.createJsonStore<
      Record<string, number>
    >(`rill:app:${org}:${project}:bookmark:recentlyUsed`, {});
  }

  public update(bookmarkId: string | undefined) {
    if (!bookmarkId) return;
    this.recentlyUsed.setter({
      ...this.recentlyUsed.value,
      [bookmarkId]: Date.now(),
    });
  }
}
