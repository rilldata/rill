import { writable } from "svelte/store";
import type { V1AnalyzedConnector } from "../../../runtime-client";

export type DataExplorerState = {
  open: boolean;
  connector: V1AnalyzedConnector | null;
};

function createDataExplorerStore() {
  const { subscribe, set } = writable<DataExplorerState>({
    open: false,
    connector: null,
  });

  return {
    subscribe,
    open: (connector: V1AnalyzedConnector) => {
      set({ open: true, connector });
    },
    close: () => {
      set({ open: false, connector: null });
    },
  };
}

export const dataExplorerStore = createDataExplorerStore();
