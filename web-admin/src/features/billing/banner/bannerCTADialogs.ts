import { writable } from "svelte/store";

export const showUpgradeDialog = writable(false);
export const upgradeDialogType = writable<"base" | "size" | "org" | "proj">(
  "base",
);
