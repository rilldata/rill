import { browser } from "$app/environment";
import { isHiddenPath } from "@rilldata/web-common/features/entity-management/file-path-utils";
import { debounce } from "@rilldata/web-common/lib/create-debouncer";
import { type Writable, writable } from "svelte/store";

export interface DirectoryState {
  [directoryPath: string]: boolean;
}

interface CustomWritable<T> extends Writable<T> {
  setProjectScope: (scopeId: string) => void;
  expand: (directoryPath: string) => void;
  collapse: (directoryPath: string) => void;
  expandAll: (directoryPaths: string[]) => void;
  collapseAll: (directoryPaths: string[]) => void;
  toggle: (directoryPath: string) => void;
  reset: () => void;
}

// Directories are expanded by default, except hidden (dot-prefixed) directories
// and everything inside them, which start collapsed. Only paths the user has
// explicitly collapsed (or expanded) are recorded, so a freshly initialized
// project starts from these defaults and the user's changes are remembered from
// there.
const DEFAULT_STATE: DirectoryState = { "/": true };
const STORAGE_PREFIX = "file-explorer-directory-state";

// Returns whether a directory is expanded, falling back to its default when the
// user has not explicitly expanded or collapsed it.
export function isDirectoryExpanded(
  state: DirectoryState,
  path: string,
): boolean {
  return state[path] ?? !isHiddenPath(path);
}

const createDirectoryStore = (): CustomWritable<DirectoryState> => {
  const { subscribe, set, update } = writable<DirectoryState>({
    ...DEFAULT_STATE,
  });

  // The persisted state is scoped per project via setProjectScope(): the
  // instanceId is always "default" in Rill Developer, so different projects
  // opened in the same browser would otherwise share one collapsed/expanded map.
  // localStorageStore can't be reused here because its key is fixed at creation,
  // whereas the active project (and thus the key) is only known once the runtime
  // metadata resolves.
  let storageKey: string | null = null;
  const persist = debounce((state: DirectoryState) => {
    if (!storageKey) return;
    try {
      localStorage.setItem(storageKey, JSON.stringify(state));
    } catch {
      // ignore: localStorage unavailable or over quota (e.g. embed iframe)
    }
  }, 300);
  subscribe((state) => persist(state));

  return {
    subscribe,
    set,
    update,
    setProjectScope: (scopeId: string) => {
      const key = `${STORAGE_PREFIX}::${scopeId}`;
      if (key === storageKey) return;
      storageKey = key;
      if (!browser) return;
      try {
        // Accessing localStorage can throw (e.g. a storage-partitioned embed
        // iframe raises a SecurityError), so fall back to the default state.
        const stored = localStorage.getItem(key);
        set(stored ? JSON.parse(stored) : { ...DEFAULT_STATE });
      } catch {
        set({ ...DEFAULT_STATE });
      }
    },
    expand: (directoryPath: string) => {
      update((state) => {
        const newState = { ...state };

        const paths = directoryPath.split("/");
        let currentPath = "";

        // Expand all directories in the path (including any parent directories)
        for (const segment of paths) {
          if (segment === "") continue;
          currentPath = currentPath + "/" + segment;
          newState[currentPath] = true;
        }

        return newState;
      });
    },
    collapse: (directoryPath: string) => {
      update((state) => ({ ...state, [directoryPath]: false }));
    },
    expandAll: (directoryPaths: string[]) => {
      update((state) => {
        const newState = { ...state };
        for (const path of directoryPaths) {
          newState[path] = true;
        }
        return newState;
      });
    },
    collapseAll: (directoryPaths: string[]) => {
      update((state) => {
        const newState = { ...state };
        for (const path of directoryPaths) {
          // The root is not collapsible; collapsing it would hide the whole tree.
          if (path === "/") continue;
          newState[path] = false;
        }
        return newState;
      });
    },
    toggle: (directoryPath: string) => {
      update((state) => ({
        ...state,
        [directoryPath]: !isDirectoryExpanded(state, directoryPath),
      }));
    },
    reset: () => {
      set({ ...DEFAULT_STATE });
    },
  };
};

export const directoryState = createDirectoryStore();
