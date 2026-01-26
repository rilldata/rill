import { writable } from "svelte/store";

interface CellInspectorState {
  isOpen: boolean;
  value: string;
}

function createCellInspectorStore() {
  const { subscribe, update } = writable<CellInspectorState>({
    isOpen: false,
    value: "",
  });

  return {
    subscribe,
    open: (value: string) =>
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
    updateValue: (value: string) =>
      update((state) => ({
        ...state,
        value,
      })),
    toggle: (value: string) =>
      update((state) => ({
        ...state,
        isOpen: !state.isOpen,
        // When opening: prefer store's existing value (from hover) if set, fall back to passed value
        // When closing: keep the current value
        value: state.isOpen ? state.value : state.value || value,
      })),
  };
}

export const cellInspectorStore = createCellInspectorStore();
