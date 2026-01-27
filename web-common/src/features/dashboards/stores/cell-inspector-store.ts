import { writable } from "svelte/store";

interface CellInspectorState {
  isOpen: boolean;
  hasValue: boolean;
  value: string | null;
}

/**
 * Converts a raw cell value to the format expected by the store.
 * Returns null for null/undefined values, string for everything else.
 */
function normalizeValue(value: unknown): string | null {
  if (value === null || value === undefined) {
    return null;
  }
  return String(value);
}

function createCellInspectorStore() {
  const { subscribe, update } = writable<CellInspectorState>({
    isOpen: false,
    hasValue: false,
    value: null,
  });

  return {
    subscribe,
    open: (value: string | null) =>
      update((state) => ({
        ...state,
        isOpen: true,
        hasValue: true,
        value,
      })),
    close: () =>
      update((state) => ({
        ...state,
        isOpen: false,
      })),
    /**
     * Update the value without changing visibility.
     * Accepts any value type and normalizes it internally.
     */
    updateValue: (value: unknown) =>
      update((state) => ({
        ...state,
        hasValue: true,
        value: normalizeValue(value),
      })),
    toggle: (value: string | null) =>
      update((state) => ({
        ...state,
        isOpen: !state.isOpen,
        // When opening: prefer store's existing value (from hover) if set, fall back to passed value
        // When closing: keep the current value
        ...(state.isOpen
          ? {}
          : {
              hasValue: true,
              value: state.hasValue ? state.value : value,
            }),
      })),
  };
}

export const cellInspectorStore = createCellInspectorStore();
