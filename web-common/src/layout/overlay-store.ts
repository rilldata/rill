import type { SvelteComponent } from "svelte";
import { writable } from "svelte/store";

export const importOverlayVisible = writable(false);

interface Overlay {
  title: string;
  detail?: {
    component: typeof SvelteComponent;
    props: Record<string, unknown>;
  };
}

export const overlay = writable<Overlay | null>(null);
