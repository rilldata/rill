import { type Writable, writable } from "svelte/store";

interface ResourceSectionState {
  [sectionName: string]: boolean;
}

interface CustomWritable<T> extends Writable<T> {
  expand: (sectionName: string) => void;
  collapse: (sectionName: string) => void;
  toggle: (sectionName: string) => void;
  reset: () => void;
}

const createResourceSectionStore = (): CustomWritable<ResourceSectionState> => {
  const { subscribe, set, update } = writable<ResourceSectionState>({
    metrics: true,
    models: true,
    dashboards: true,
  });

  return {
    subscribe,
    set,
    update,
    expand: (sectionName: string) => {
      update((state) => ({ ...state, [sectionName]: true }));
    },
    collapse: (sectionName: string) => {
      update((state) => ({ ...state, [sectionName]: false }));
    },
    toggle: (sectionName: string) => {
      update((state) => ({ ...state, [sectionName]: !state[sectionName] }));
    },
    reset: () => {
      set({
        metrics: true,
        models: true,
        dashboards: true,
      });
    },
  };
};

export const resourceSectionState = createResourceSectionStore();
