import { writable } from "svelte/store";

interface DbtImportModalState {
  open: boolean;
  connectorName: string;
}

function createDbtImportModalStore() {
  const { subscribe, set } = writable<DbtImportModalState>({
    open: false,
    connectorName: "",
  });

  return {
    subscribe,
    open(connectorName: string) {
      set({ open: true, connectorName });
    },
    close() {
      set({ open: false, connectorName: "" });
    },
  };
}

export const dbtImportModal = createDbtImportModalStore();
