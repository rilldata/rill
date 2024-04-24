import { Writable, writable } from "svelte/store";

interface DirectoryState {
  [directoryPath: string]: boolean;
}

interface CustomWritable<T> extends Writable<T> {
  expand: (directoryPath: string) => void;
  collapse: (directoryPath: string) => void;
  toggle: (directoryPath: string) => void;
  reset: () => void;
}

const createDirectoryStore = (): CustomWritable<DirectoryState> => {
  const { subscribe, set, update } = writable<DirectoryState>({
    "/": true,
  });

  return {
    subscribe,
    set,
    update,
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
    toggle: (directoryPath: string) => {
      update((state) => ({ ...state, [directoryPath]: !state[directoryPath] }));
    },
    reset: () => {
      set({ "/": true });
    },
  };
};

export const directoryState = createDirectoryStore();
