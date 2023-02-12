import { writable } from "svelte/store";

export const importOverlayVisible = writable(false);

interface Overlay {
  title: string;
  message?: string;
}

export const overlay = writable<Overlay>(null);
