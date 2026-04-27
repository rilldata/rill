import { writable } from "svelte/store";
import type { SearchableItem } from "./types";

const STORAGE_KEY = "rill:command-palette:recents";
const MAX_RECENTS = 8;

function loadRecents(): SearchableItem[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    return JSON.parse(raw) as SearchableItem[];
  } catch {
    return [];
  }
}

function saveRecents(items: SearchableItem[]): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
  } catch {
    // localStorage full or unavailable; silently ignore
  }
}

function createRecentsStore() {
  const { subscribe, set, update } = writable<SearchableItem[]>(loadRecents());

  return {
    subscribe,
    /** Push an item to the front of recents, deduplicating by route. */
    add(item: SearchableItem) {
      update((items) => {
        const filtered = items.filter((i) => i.route !== item.route);
        const next = [item, ...filtered].slice(0, MAX_RECENTS);
        saveRecents(next);
        return next;
      });
    },
    /** Clear recents for a specific org (e.g. on org switch). */
    clearForOrg(orgName: string) {
      update((items) => {
        const next = items.filter((i) => i.orgName !== orgName);
        saveRecents(next);
        return next;
      });
    },
  };
}

export const recentItems = createRecentsStore();
