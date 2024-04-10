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
  const { subscribe, set, update } = writable<DirectoryState>({});

  return {
    subscribe,
    set,
    update,
    expand: (directoryPath: string) => {
      update((state) => ({ ...state, [directoryPath]: true }));
    },
    collapse: (directoryPath: string) => {
      update((state) => ({ ...state, [directoryPath]: false }));
    },
    toggle: (directoryPath: string) => {
      update((state) => ({ ...state, [directoryPath]: !state[directoryPath] }));
    },
    reset: () => {
      set({});
    },
  };
};

export const directoryState = createDirectoryStore();
