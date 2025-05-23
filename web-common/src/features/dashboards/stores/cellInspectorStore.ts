import { writable } from "svelte/store";

interface CellInspectorState {
  isOpen: boolean;
  value: string;
  position: { x: number; y: number } | null;
}

function createCellInspectorStore() {
  const { subscribe, update } = writable<CellInspectorState>({
    isOpen: false,
    value: "",
    position: null,
  });

  return {
    subscribe,
    open: (value: string, position: { x: number; y: number }) =>
      update((state) => ({
        ...state,
        isOpen: true,
        value,
        position,
      })),
    close: () =>
      update((state) => ({
        ...state,
        isOpen: false,
      })),
    toggle: (value: string, position: { x: number; y: number }) =>
      update((state) => ({
        ...state,
        isOpen: !state.isOpen,
        value: state.isOpen ? state.value : value,
        position: state.isOpen ? state.position : position,
      })),
  };
}

export const cellInspectorStore = createCellInspectorStore();
