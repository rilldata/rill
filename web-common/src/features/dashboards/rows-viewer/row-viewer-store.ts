import { writable } from "svelte/store";

export const rowViewerStore = writable<HTMLElement | null>(null);
