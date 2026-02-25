import { writable, get } from "svelte/store";
import type { V1User } from "../../client";
import { browser } from "$app/environment";

const STORAGE_KEY_PREFIX = "rill:viewAsUser:";

function getStorageKey(org: string, project: string): string {
  return `${STORAGE_KEY_PREFIX}${org}/${project}`;
}

function createViewAsUserStore() {
  const store = writable<V1User | null>(null);
  let currentScope: { org: string; project: string } | null = null;

  return {
    subscribe: store.subscribe,

    /**
     * Initialize the store for a specific project scope.
     * Loads persisted state from sessionStorage if available.
     */
    initForProject(org: string, project: string): void {
      if (!browser) return;

      // Avoid redundant sessionStorage reads when scope hasn't changed
      if (currentScope?.org === org && currentScope?.project === project) return;

      currentScope = { org, project };
      const key = getStorageKey(org, project);

      try {
        const stored = sessionStorage.getItem(key);
        if (stored) {
          const user = JSON.parse(stored) as V1User;
          store.set(user);
        } else {
          store.set(null);
        }
      } catch {
        store.set(null);
      }
    },

    /**
     * Set the view-as user for the current project scope.
     */
    set(user: V1User | null): void {
      store.set(user);

      if (!browser || !currentScope) return;

      const key = getStorageKey(currentScope.org, currentScope.project);
      try {
        if (user) {
          sessionStorage.setItem(key, JSON.stringify(user));
        } else {
          sessionStorage.removeItem(key);
        }
      } catch {
        // Ignore storage errors
      }
    },

    /**
     * Clear the view-as state (e.g., when navigating away from project).
     */
    clear(): void {
      store.set(null);

      if (!browser || !currentScope) return;

      const key = getStorageKey(currentScope.org, currentScope.project);
      try {
        sessionStorage.removeItem(key);
      } catch {
        // Ignore storage errors
      }
      currentScope = null;
    },

    /**
     * Get current value synchronously.
     */
    get(): V1User | null {
      return get(store);
    },
  };
}

export const viewAsUserStore = createViewAsUserStore();
