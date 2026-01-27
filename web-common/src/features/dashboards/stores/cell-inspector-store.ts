import { writable } from "svelte/store";

interface CellInspectorState {
  isOpen: boolean;
  // undefined = no value set yet, null = cell contains null, string = cell value (including empty string)
  value: string | null | undefined;
}

function createCellInspectorStore() {
  const { subscribe, update } = writable<CellInspectorState>({
    isOpen: false,
    value: undefined,
  });

  return {
    subscribe,
    open: (value: string | null) =>
      update((state) => ({
        ...state,
        isOpen: true,
        value,
      })),
    close: () =>
      update((state) => ({
        ...state,
        isOpen: false,
      })),
    // Update the value without changing visibility
    updateValue: (value: string | null) =>
      update((state) => ({
        ...state,
        value,
      })),
    toggle: (value: string | null) =>
      update((state) => ({
        ...state,
        isOpen: !state.isOpen,
        // When opening: prefer store's existing value (from hover) if set, fall back to passed value
        // When closing: keep the current value
        // undefined means no value was set via hover, so fall back to passed value
        value: state.isOpen
          ? state.value
          : state.value !== undefined
            ? state.value
            : value,
      })),
  };
}

export const cellInspectorStore = createCellInspectorStore();
