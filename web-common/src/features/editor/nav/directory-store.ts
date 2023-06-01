import { Writable, writable } from "svelte/store";

interface DirectoryState {
  [directoryName: string]: boolean;
}

interface CustomWritable<T> extends Writable<T> {
  expand: (directoryName: string) => void;
  collapse: (directoryName: string) => void;
  toggle: (directoryName: string) => void;
  reset: () => void;
}

const createDirectoryStore = (): CustomWritable<DirectoryState> => {
  const { subscribe, set, update } = writable<DirectoryState>({});

  return {
    subscribe,
    set,
    update,
    expand: (directoryName: string) => {
      update((state) => ({ ...state, [directoryName]: true }));
    },
    collapse: (directoryName: string) => {
      update((state) => ({ ...state, [directoryName]: false }));
    },
    toggle: (directoryName: string) => {
      update((state) => ({ ...state, [directoryName]: !state[directoryName] }));
    },
    reset: () => {
      set({});
    },
  };
};

export const directoryState = createDirectoryStore();
