import { localStorageStore } from "../lib/store-utils/local-storage";

export const previewModeStore = localStorageStore<boolean>(
  "rill:preview-mode",
  false,
);
