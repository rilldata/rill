import { derived, Readable, writable, Writable } from "svelte/store";

export enum DuplicateActions {
  None = "NONE",
  KeepBoth = "KEEP_BOTH",
  Overwrite = "OVERWRITE",
  Cancel = "CANCEL",
}

export const duplicateSourceAction: Writable<DuplicateActions> = writable(
  DuplicateActions.None
);

export const duplicateSourceName: Writable<string> = writable(null);

interface Source {
  clientYAML: string;
  setClientYAML: (yaml: string) => void;
}

interface SourcesStore {
  [name: string]: Source;
}

const sourcesStore = writable<SourcesStore>({});

// TODO: clean up
export function useSourceStore(name: string): Readable<Source> {
  const source: Source = {
    clientYAML: "",
    setClientYAML(yaml: string) {
      source.clientYAML = yaml;
      sourcesStore.update((state) => ({
        ...state,
        [name]: {
          ...state[name],
          clientYAML: yaml,
        },
      }));
    },
  };

  sourcesStore.update((state) => {
    if (!state[name]) {
      state[name] = source;
    }
    return state;
  });

  return derived(sourcesStore, ($sourcesStore: SourcesStore) => ({
    ...$sourcesStore[name],
    setClientYAML: source.setClientYAML,
  }));
}
